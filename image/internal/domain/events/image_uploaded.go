package events

import (
	"time"

	"github.com/google/uuid"
)

type ImageUploaded struct {
	EventID     uuid.UUID `json:"event_id"`
	ImageID     uuid.UUID `json:"image_id"`
	UserID      uuid.UUID `json:"user_id"`
	StorageKey  string    `json:"storage_key"`
	OriginalName string   `json:"original_name"`
	MimeType    string    `json:"mime_type"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	FileSize    int64     `json:"file_size"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func NewImageUploaded(imageID, userID uuid.UUID, storageKey, originalName, mimeType string, width, height int, fileSize int64) ImageUploaded {
	return ImageUploaded{
		EventID:      uuid.New(),
		ImageID:      imageID,
		UserID:       userID,
		StorageKey:   storageKey,
		OriginalName: originalName,
		MimeType:     mimeType,
		Width:        width,
		Height:       height,
		FileSize:     fileSize,
		OccurredAt:   time.Now(),
	}
}
