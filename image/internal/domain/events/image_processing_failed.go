package events

import (
	"time"

	"github.com/google/uuid"
)

type ImageProcessingFailed struct {
	EventID    uuid.UUID `json:"event_id"`
	ImageID    uuid.UUID `json:"image_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}
