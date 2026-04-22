package events

import "github.com/google/uuid"

type UserUnbannedEvent struct {
	Version int       `json:"version"`
	ID      uuid.UUID `json:"id"`
}
