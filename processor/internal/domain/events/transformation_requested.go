package events

import (
	"time"

	"github.com/google/uuid"
)

const EventTypeTransformationRequested = "TransformationRequested"

type TransformationRequested struct {
	EventID          uuid.UUID `json:"event_id"`
	TransformationID uuid.UUID `json:"transformation_id"`
	OccurredAt       time.Time `json:"occurred_at"`
}

func NewTransformationRequested(transformationID uuid.UUID) TransformationRequested {
	return TransformationRequested{
		EventID:          uuid.New(),
		TransformationID: transformationID,
		OccurredAt:       time.Now().UTC(),
	}
}
