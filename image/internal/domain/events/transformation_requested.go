package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const EventTypeTransformationRequested = "TransformationRequested"

type TransformationRequested struct {
	EventID       uuid.UUID        `json:"event_id"`
	ImageID       uuid.UUID        `json:"image_id"`
	TransformID   uuid.UUID        `json:"transform_id"`
	TransformSpec json.RawMessage  `json:"transform_spec"`
	OccurredAt    time.Time        `json:"occurred_at"`
}

func NewTransformationRequested(imageID, transformID uuid.UUID, spec json.RawMessage) TransformationRequested {
	return TransformationRequested{
		EventID:       uuid.New(),
		ImageID:       imageID,
		TransformID:   transformID,
		TransformSpec: spec,
		OccurredAt:    time.Now(),
	}
}
