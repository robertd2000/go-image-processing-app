package model

import (
	"io"

	"github.com/google/uuid"
)

type UploadImageInput struct {
	UserID      uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}
