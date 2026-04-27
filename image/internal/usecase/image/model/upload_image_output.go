package model

import (
	"time"

	"github.com/google/uuid"
)

type UploadImageOutput struct {
	ImageID   uuid.UUID
	CreatedAt time.Time
}
