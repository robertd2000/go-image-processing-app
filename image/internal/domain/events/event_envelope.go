package events

import "encoding/json"

const (
	EventTypeImageUploaded       = "ImageUploaded"
	EventTypeImageProcessingCompleted = "ImageProcessingCompleted"
	EventTypeImageProcessingFailed    = "ImageProcessingFailed"
)

type EventEnvelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}
