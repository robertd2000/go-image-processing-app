package events

import (
	"time"

	"github.com/google/uuid"
)

type ImageProcessingCompleted struct {
	EventID    uuid.UUID `json:"event_id"`
	ImageID    uuid.UUID `json:"image_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
