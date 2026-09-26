package events

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the unique discriminator for a domain event.
type EventType string

// BaseEvent provides standard metadata for all events across the platform.
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// NewBaseEvent constructs a BaseEvent with a generated UUID and current UTC timestamp.
func NewBaseEvent(eventType EventType) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
	}
}

// Event defines the contract that all domain events must satisfy.
type Event interface {
	GetID() string
	GetType() EventType
	GetTimestamp() time.Time
}

// GetID returns the unique identifier of the event.
func (e BaseEvent) GetID() string {
	return e.ID
}

// GetType returns the event type discriminator.
func (e BaseEvent) GetType() EventType {
	return e.Type
}

// GetTimestamp returns the UTC creation time of the event.
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}
