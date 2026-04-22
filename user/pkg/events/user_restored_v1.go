package events

import "github.com/google/uuid"

type UserRestoredEvent struct {
	ID uuid.UUID `json:"id"`
}
