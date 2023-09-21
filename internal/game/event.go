package game

import (
	"encoding/json"
	"fmt"
)

const (
	Heartbeat    EventType = "heartbeat"
	Registration           = "registration"
	EventRound             = "round"
	EventShot                   = "shot"
)

type EventType string

type Event struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func NewEvent(eventType EventType, data interface{}) (*Event, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	return &Event{
		Type: eventType,
		Data: payload,
	}, nil
}
