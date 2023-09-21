package game

import (
	"encoding/json"
	"fmt"
	"github.com/reactivejson/cowboys/internal/domain"
	"log"
	"sync"
)

// Various error messages
var (
	ErrGameNotStarted            = fmt.Errorf("game not gameStarted yet")
	ErrGameAlreadyStarted        = fmt.Errorf("game already started")
	ErrInvalidPayload            = fmt.Errorf("invalid payload")
	ErrGameFinished              = fmt.Errorf("game is over")
	ErrInvalidPlayerRegistration = fmt.Errorf("invalid player registration event")
)

type Game struct {
	gameStarted, gameFinished bool
	totalPlayers              int

	players map[string]*domain.Player
	lock    *sync.Mutex
}

// NewGame creates a new game state based on the provided configuration.

func NewGame(cfg *domain.MasterConfig) *Game {
	return &Game{
		totalPlayers: cfg.Players,
		players:      make(map[string]*domain.Player),
		lock:         new(sync.Mutex),
	}
}

// EmitEvent generates an event based on the current game state.
func (gs *Game) EmitEvent() (*Event, error) {
	gs.lock.Lock()
	defer gs.lock.Unlock()

	if gs.gameFinished {
		return nil, ErrGameFinished
	}

	// If the game hasn't started yet, emit a heartbeat event.
	if !gs.gameStarted {
		return NewEvent(Heartbeat, nil)
	}

	if len(gs.players) == 1 {
		gs.gameFinished = true
	}
	// Emit a round event with player information.
	return NewEvent(EventRound, &domain.Round{
		Players: gs.players,
	})
}

// HandleEvent processes incoming events and updates the game state accordingly.
func (gs *Game) HandleEvent(event *Event) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()

	switch event.Type {
	case Registration:
		return gs.handlePlayerRegistration(event)
	case EventShot:
		return gs.handlePlayerAction(event)
		// Ignore unsupported events.
	default:
		return nil
	}
}

// handlePlayerRegistration processes a player registration event and adds players to the game.
func (gs *Game) handlePlayerRegistration(event *Event) error {
	if gs.gameStarted {
		return ErrGameAlreadyStarted
	}

	var player domain.Player
	if err := json.Unmarshal(event.Data, &player); err != nil {
		return fmt.Errorf("failed to unmarshal player registration payload: %w", err)
	}

	if player.IsEmpty() {
		return ErrInvalidPlayerRegistration
	}

	gs.players[player.ID] = &player

	if len(gs.players) == gs.totalPlayers {
		gs.gameStarted = true
	}

	return nil
}

// handlePlayerAction processes a player action event and updates player status.
func (gs *Game) handlePlayerAction(event *Event) error {
	if !gs.gameStarted {
		return ErrGameNotStarted
	}

	var action domain.Action
	if err := json.Unmarshal(event.Data, &action); err != nil {
		return fmt.Errorf("failed to unmarshal player action payload: %w", err)
	}

	if action.Src == "" || action.Dest == "" {
		return ErrInvalidPayload
	}

	// Check if the 'action.Src' and 'action.Dest' players exist.
	fromPlayer, fromExists := gs.players[action.Src]
	toPlayer, toExists := gs.players[action.Dest]

	if !fromExists || !toExists {
		// One or both players no longer exist.
		return nil
	}

	// Apply the action on the target player.
	toPlayer.Health -= fromPlayer.Damage

	log.Printf(
		"%s Action %d damage on %s",
		fromPlayer.Name,
		fromPlayer.Damage,
		toPlayer.Name,
	)

	if toPlayer.Health < 1 {
		// Remove the defeated player from the game.
		delete(gs.players, action.Dest)
	}

	return nil
}
