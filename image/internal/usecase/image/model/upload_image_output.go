package model

import (
	"time"

	"github.com/google/uuid"
)

type UploadImageOutput struct {
	ImageID   uuid.UUID
	Status    string
	CreatedAt time.Time
}
