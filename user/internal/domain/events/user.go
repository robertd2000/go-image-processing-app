// Package events
package domainevents

import (
	"time"

	"github.com/google/uuid"
)

type UserDeletedEvent struct {
	ID        uuid.UUID
	DeletedAt time.Time
}

type UserBannedEvent struct {
	ID     uuid.UUID
	Reason string
}

type UserUnbannedEvent struct {
	ID uuid.UUID
}

type UserRestoredEvent struct {
	ID uuid.UUID
}

type UserCreatedEvent struct {
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}
