package events

import (
	"time"

	"github.com/google/uuid"
)

type UserDeletedEvent struct {
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}
