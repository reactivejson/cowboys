package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/reactivejson/cowboys/internal/domain"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/reactivejson/cowboys/internal/game"
)

const (
	masterTopic  = "master_events"
	registerPath = "/join"
)

type registrationRequest struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

type Master struct {
	cfg           *domain.MasterConfig
	ctx           context.Context
	cancel        context.CancelFunc
	state         *game.Game
	logger        *log.Logger
	redisClient   *redis.Client
	lastRoundData json.RawMessage
}

func NewMaster(cfg *domain.MasterConfig, state *game.Game, logger *log.Logger, redisClient *redis.Client) *Master {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	return &Master{
		ctx:         ctx,
		cancel:      cancel,
		cfg:         cfg,
		state:       state,
		logger:      logger,
		redisClient: redisClient,
	}
}

func (m *Master) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc(registerPath, m.handleRegistration)

	server := &http.Server{
		Addr:    m.cfg.Port,
		Handler: mux,
	}

	ticker := time.Tick(time.Second)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			m.logger.Printf("master HTTP server listen: %v", err)
		}
	}()

	go func() {
		subscription := m.redisClient.Subscribe(m.ctx, playerTopic)
		for {
			select {
			case msg := <-subscription.Channel():
				m.handleMessage(msg)
			case <-m.ctx.Done():
				if err := subscription.Close(); err != nil {
					m.logger.Printf("close competitor events channel: %v", err)
				}

				return
			}
		}
	}()

	for {
		select {
		case <-ticker:
			m.beat()
		case <-m.ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.TODO(), time.Minute)
			if err := server.Shutdown(shutdownCtx); err != nil {
				m.logger.Printf("master HTTP server shutdown: %v", err)
			}

			cancel()
			return
		}
	}
}

func (m *Master) handleMessage(msg *redis.Message) {
	var event game.Event
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		m.logger.Printf("unmarshal competitor event: %v", err)
		m.cancel()
		return
	}

	err := m.state.HandleEvent(&event)
	if err != nil && err != game.ErrInvalidPayload {
		m.logger.Printf("handle competitor event: %v", err)
		m.cancel()
		return
	}

	if err != nil {
		m.logger.Printf("competitor event: %v", err)
	}
}

func (m *Master) beat() {
	event, err := m.state.EmitEvent()
	if err != nil {
		m.logger.Printf("emit event: %v", err)
		m.cancel()
		return
	}

	if event.Type == game.EventRound {
		defer func() { m.lastRoundData = event.Data }()
	}

	if bytes.Equal(event.Data, m.lastRoundData) {
		m.logger.Printf("state did not change, not enough competitors")
		m.cancel()
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		m.logger.Printf("marshal event: %v", err)
		m.cancel()
		return
	}

	if err := m.redisClient.Publish(m.ctx, masterTopic, payload).Err(); err != nil {
		m.logger.Printf("publish event: %v", err)
		m.cancel()
		return
	}
}

func (m *Master) handleRegistration(w http.ResponseWriter, r *http.Request) {
	var request registrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		m.logger.Printf("decode request body: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if request.Name == "" || request.Health == 0 || request.Damage == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	player := domain.Player{
		ID:     uuid.NewString(),
		Name:   request.Name,
		Health: request.Health,
		Damage: request.Damage,
	}

	event, err := game.NewEvent(game.Registration, &player)
	if err != nil {
		m.logger.Printf("create registration event: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = m.state.HandleEvent(event); err != nil {
		if err == game.ErrGameFinished || err == game.ErrGameAlreadyStarted {
			http.Error(w, "conflict", http.StatusConflict)
			return
		}

		m.logger.Printf("create registration event: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(player); err != nil {
		m.logger.Printf("encode registration response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
