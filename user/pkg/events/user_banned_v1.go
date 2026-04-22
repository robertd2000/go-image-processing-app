package events

import "github.com/google/uuid"

type UserBannedEvent struct {
	Version int       `json:"version"`
	ID      uuid.UUID `json:"id"`
}
