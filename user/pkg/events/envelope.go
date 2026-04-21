package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID    uuid.UUID       `json:"event_id"`
	EventType  string          `json:"event_type"`
	Version    int             `json:"version"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// helpers
func NewEvent[T any](eventType string, version int, payload T) (Event, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}

	return Event{
		EventID:    uuid.New(),
		EventType:  eventType,
		Version:    version,
		OccurredAt: time.Now(),
		Payload:    raw,
	}, nil
}

func ParsePayload[T any](evt Event, out *T) error {
	return json.Unmarshal(evt.Payload, out)
}
