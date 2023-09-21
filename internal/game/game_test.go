package game

import (
	"encoding/json"
	"testing"

	"github.com/reactivejson/cowboys/internal/domain"
)

func TestGame(t *testing.T) {
	state := NewGame(&domain.MasterConfig{
		Players: 2,
	})

	event, err := state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected first event emission err: %v", err)
	}

	if event.Type != Heartbeat {
		t.Fatalf("first event emission should be of type %q, got %q", Heartbeat, event.Type)
	}

	firstRegistration, _ := NewEvent(Registration, &domain.Player{
		ID:     "test_1",
		Name:   "Test1",
		Health: 3,
		Damage: 1,
	})
	if err := state.HandleEvent(firstRegistration); err != nil {
		t.Fatalf("unexpected first registration err: %v", err)
	}

	secondRegistration, _ := NewEvent(Registration, &domain.Player{
		ID:     "test_2",
		Name:   "Test2",
		Health: 1,
		Damage: 1,
	})
	if err := state.HandleEvent(secondRegistration); err != nil {
		t.Fatalf("unexpected second registration err: %v", err)
	}

	event, err = state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected second emission err: %v", err)
	}

	if event.Type != EventRound {
		t.Fatalf("expected second emission event type to be %q, got %q", EventRound, event.Type)
	}

	firstShotEvent, _ := NewEvent(EventShot, &domain.Action{
		Src:  "test_1",
		Dest: "test_2",
	})
	if err := state.HandleEvent(firstShotEvent); err != nil {
		t.Fatalf("unexpected first shot err: %v", err)
	}

	secondShot, _ := NewEvent(EventShot, &domain.Action{
		Src:  "test_2",
		Dest: "test_1",
	})
	if err := state.HandleEvent(secondShot); err != nil {
		t.Fatalf("unexpected second shot err: %v", err)
	}

	event, err = state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected third emission err: %v", err)
	}

	var round domain.Round
	if err := json.Unmarshal(event.Data, &round); err != nil {
		t.Fatalf("can not unmarshal round event: %v", err)
	}

	if len(round.Players) != 1 {
		t.Fatalf("expected to receive round event with 1 competitor, got %d", len(round.Players))
	}

	event, err = state.EmitEvent()
	if err != ErrGameFinished {
		t.Fatalf("fourth emission expected to be ErrGameFinished, got: %v", err)
	}
}

func TestGameJoin(t *testing.T) {
	state := NewGame(&domain.MasterConfig{
		Players: 1,
	})

	event, err := state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected first event emission err: %v", err)
	}

	if event.Type != Heartbeat {
		t.Fatalf("first event emission should be of type %q, got %q", Heartbeat, event.Type)
	}

	firstRegistration, _ := NewEvent(Registration, &domain.Player{
		ID:     "test_1",
		Name:   "Test1",
		Health: 3,
		Damage: 1,
	})
	if err := state.HandleEvent(firstRegistration); err != nil {
		t.Fatalf("unexpected first registration err: %v", err)
	}

	event, err = state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected second emission err: %v", err)
	}

	if event.Type != EventRound {
		t.Fatalf("expected second emission event type to be %q, got %q", EventRound, event.Type)
	}

	secondRegistration, err := NewEvent(Registration, &domain.Player{
		ID:     "test_2",
		Name:   "Test2",
		Health: 1,
		Damage: 1,
	})
	if err := state.HandleEvent(secondRegistration); err != ErrGameAlreadyStarted {
		t.Fatalf("unexpected second registration err: %v", err)
	}
}

func TestGameInvalidRegistration(t *testing.T) {
	state := NewGame(&domain.MasterConfig{
		Players: 1,
	})

	event, err := state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected first event emission err: %v", err)
	}

	if event.Type != Heartbeat {
		t.Fatalf("first event emission should be of type %q, got %q", Heartbeat, event.Type)
	}

	firstRegistration, _ := NewEvent(Registration, &domain.Player{})
	if err := state.HandleEvent(firstRegistration); err != ErrInvalidPlayerRegistration {
		t.Fatalf("unexpected first registration err: %v", err)
	}
}

func TestGameActionEvent(t *testing.T) {
	state := NewGame(&domain.MasterConfig{
		Players: 2,
	})

	event, err := state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected first event emission err: %v", err)
	}

	if event.Type != Heartbeat {
		t.Fatalf("first event emission should be of type %q, got %q", Heartbeat, event.Type)
	}

	shotEvent, _ := NewEvent(EventShot, &domain.Action{})
	if err := state.HandleEvent(shotEvent); err != ErrGameNotStarted {
		t.Fatalf("expected not gameStarted error, got: %v", err)
	}

	firstRegistration, _ := NewEvent(Registration, &domain.Player{
		ID:     "test_1",
		Name:   "Test1",
		Health: 3,
		Damage: 1,
	})
	if err := state.HandleEvent(firstRegistration); err != nil {
		t.Fatalf("unexpected first registration err: %v", err)
	}

	secondRegistration, _ := NewEvent(Registration, &domain.Player{
		ID:     "test_2",
		Name:   "Test2",
		Health: 1,
		Damage: 1,
	})
	if err := state.HandleEvent(secondRegistration); err != nil {
		t.Fatalf("unexpected second registration err: %v", err)
	}

	shotEvent, _ = NewEvent(EventShot, &domain.Action{})
	if err := state.HandleEvent(shotEvent); err != ErrInvalidPayload {
		t.Fatalf("expected unacceptable payload eror, got: %v", err)
	}

	shotEvent, _ = NewEvent(EventShot, &domain.Action{
		Src:  "unexisting",
		Dest: "test_1",
	})
	if err := state.HandleEvent(shotEvent); err != nil {
		t.Fatalf("unexpected error when handling shot from unexisting player: %v", err)
	}

	event, err = state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected err when emiting event after shot: %v", err)
	}

	if event.Type != EventRound {
		t.Fatalf("expected event type %q, got %q", EventRound, event.Type)
	}

	var round domain.Round
	if err := json.Unmarshal(event.Data, &round); err != nil {
		t.Fatalf("unexpected error when unmarshaling round event: %v", err)
	}

	if round.Players["test_1"].Health != 3 || round.Players["test_2"].Health != 1 {
		t.Fatalf("unexpected change in players' health")
	}

	shotEvent, _ = NewEvent(EventShot, &domain.Action{
		Src:  "test_2",
		Dest: "test_1",
	})
	if err := state.HandleEvent(shotEvent); err != nil {
		t.Fatalf("unexpected error when handling proper shot: %v", err)
	}

	event, err = state.EmitEvent()
	if err != nil {
		t.Fatalf("unexpected err when emiting event after shot: %v", err)
	}

	var round2 domain.Round
	if err := json.Unmarshal(event.Data, &round2); err != nil {
		t.Fatalf("unexpected error when unmarshaling round event: %v", err)
	}

	if round2.Players["test_1"].Health != 2 || round2.Players["test_2"].Health != 1 {
		t.Fatalf("incorrect change in players' health")
	}

	shotEvent, _ = NewEvent(EventShot, &domain.Action{
		Src:  "test_2",
		Dest: "nonexisting",
	})
	if err := state.HandleEvent(shotEvent); err != nil {
		t.Fatalf("unexpected error when handling shot to nonexisting target: %v", err)
	}

	var round3 domain.Round
	if err := json.Unmarshal(event.Data, &round3); err != nil {
		t.Fatalf("unexpected error when unmarshaling round event: %v", err)
	}

	if round2.Players["test_1"].Health != 2 || round2.Players["test_2"].Health != 1 {
		t.Fatalf("unexpected change in players' health")
	}
}
