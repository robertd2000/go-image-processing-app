package domainevents

import (
	"time"

	"github.com/google/uuid"
)

type UserDeletedEvent struct {
	Version   int
	ID        uuid.UUID
	DeletedAt time.Time
}

type UserBannedEvent struct {
	ID     uuid.UUID `json:"id"`
	Reason string    `json:"reason"`
}
type UserCreatedEvent struct {
	Version   int
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}
