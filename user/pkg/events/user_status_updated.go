package events

import (
	"time"

	"github.com/google/uuid"
)

type UserStatusUpdatedEvent struct {
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
