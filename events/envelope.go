package events

import (
	"time"

	"github.com/google/uuid"
)

type Event[T any] struct {
	EventID    uuid.UUID `json:"event_id"`
	EventType  string    `json:"event_type"`
	Version    int       `json:"version"`
	OccurredAt time.Time `json:"occurred_at"`
	Payload    T         `json:"payload"`
}
