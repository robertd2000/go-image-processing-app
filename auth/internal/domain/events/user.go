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

type UserCreatedEvent struct {
	Version   int
	ID        uuid.UUID
	Username  string
	Email     string
	CreatedAt time.Time
}
