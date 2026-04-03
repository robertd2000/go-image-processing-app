package events

import (
	"time"

	"github.com/google/uuid"
)

type UserCreatedEvent struct {
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
