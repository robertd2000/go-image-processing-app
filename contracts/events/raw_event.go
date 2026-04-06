package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RawEvent struct {
	EventID    uuid.UUID       `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}
