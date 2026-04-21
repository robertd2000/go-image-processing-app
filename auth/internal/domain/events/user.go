package domainevents

import (
	"time"

	"github.com/google/uuid"
)

type UserDeletedEvent struct {
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}

type UserCreatedEvent struct {
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
