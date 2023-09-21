package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/reactivejson/cowboys/internal/domain"
	"github.com/reactivejson/cowboys/internal/game"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	playerTopic = "player_events"
)

var (
	ErrUnexpectedEvent = fmt.Errorf("unexpected event received")
	ErrNoTarget        = fmt.Errorf("no target found")
)

type Player struct {
	ID          string
	cfg         *domain.PlayerConfig
	ctx         context.Context
	cancel      context.CancelFunc
	shotChan    chan *domain.Action
	redisClient *redis.Client
	logger      *log.Logger
}

func NewPlayer(cfg *domain.PlayerConfig, redisClient *redis.Client, logger *log.Logger) *Player {
	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt)

	return &Player{
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancelFn,
		shotChan:    make(chan *domain.Action),
		redisClient: redisClient,
		logger:      logger,
	}
}

func (p *Player) Run() {
	go p.fetchActions()

	sub := p.redisClient.Subscribe(p.ctx, masterTopic)

	for {
		select {
		// finished
		case <-p.ctx.Done():
			if err := sub.Close(); err != nil {
				p.logger.Printf("close master pub/sub: %v", err)
			}

			close(p.shotChan)

			return
		// we expect to receive a message every second
		case msg := <-sub.Channel():
			if err := p.handleMasterMessage(msg); err != nil {
				p.logger.Printf("handle message from master: %v", err)
				p.cancel()
			}
		// communication is lost
		case <-time.After(time.Second * 2):
			p.logger.Printf("no heartbeat")
			p.cancel()
		}
	}
}

func (p *Player) handleMasterMessage(msg *redis.Message) error {
	var event game.Event
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch event.Type {
	case game.Heartbeat:
		if p.ID == "" {
			return p.join()
		}

		return nil
	case game.EventRound:
		if p.ID == "" {
			return ErrUnexpectedEvent
		}

		var round domain.Round
		if err := json.Unmarshal(event.Data, &round); err != nil {
			return fmt.Errorf("unmarshal competitors: %w", err)
		}

		win, ok := round.Players[p.ID]
		if len(round.Players) == 1 && ok {
			log.Println("I am the Winner:) ", win.Name, "My health", win.Health)
			p.cancel()
			return nil
		}

		if !ok {
			log.Println("They Killed me -> DEAD :(")
			p.cancel()
			return nil
		}

		for target := range round.Players {
			if target != p.ID {
				p.shotChan <- &domain.Action{
					Src:  p.ID,
					Dest: target,
				}

				return nil
			}
		}

		return ErrNoTarget
	default:
		return fmt.Errorf("unknown event %q received", event.Type)
	}
}

func (p *Player) join() error {
	masterURL, err := url.Parse(p.cfg.MasterAddr)
	if err != nil {
		return fmt.Errorf("parse master url: %w", err)
	}

	masterURL.Path = registerPath

	payload, err := json.Marshal(&registrationRequest{
		Name:   p.cfg.Name,
		Health: p.cfg.Health,
		Damage: p.cfg.Damage,
	})
	if err != nil {
		return fmt.Errorf("marshal registration request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, masterURL.String(), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create registration request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send registration request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected registration response code %d", resp.StatusCode)
	}

	var competitor domain.Player
	if err := json.NewDecoder(resp.Body).Decode(&competitor); err != nil {
		return fmt.Errorf("decode registration response: %w", err)
	}
	defer resp.Body.Close()

	p.ID = competitor.ID

	return nil
}

func (p *Player) fetchActions() {
	for shot := range p.shotChan {
		event, err := game.NewEvent(game.EventShot, shot)
		if err != nil {
			p.logger.Printf("create shot event: %v", err)
			p.cancel()
			return
		}

		payload, err := json.Marshal(event)
		if err != nil {
			p.logger.Printf("marshal shot event: %v", err)
			p.cancel()
			return
		}

		if err := p.redisClient.Publish(p.ctx, playerTopic, payload).Err(); err != nil {
			p.logger.Printf("publish shot event: %v", err)
			p.cancel()
			return
		}
	}
}
