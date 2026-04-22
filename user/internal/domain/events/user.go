// Package events
package domainevents

import (
	"time"

	"github.com/google/uuid"
)

type UserDeletedEvent struct {
	ID        uuid.UUID `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type UserBannedEvent struct {
	ID       uuid.UUID `json:"id"`
	Reason   string    `json:"reason"`
	BannedAt time.Time `json:"banned_at"`
}

type UserRestoredEvent struct {
	ID         uuid.UUID `json:"id"`
	RestoredAt time.Time `json:"restored_at"`
}

type UserCreatedEvent struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
