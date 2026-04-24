package image

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StorageKey string

func NewStorageKey(userID, imageID uuid.UUID, ext string) StorageKey {
	return StorageKey(fmt.Sprintf("originals/%s/%s.%s", userID, imageID, ext))
}

type Image struct {
	id         uuid.UUID
	userID     uuid.UUID
	storageKey StorageKey
	metadata   ImageMetadata
	createdAt  time.Time
}

func NewImage(
	userID uuid.UUID,
	originalName string,
	meta ImageMetadata,
	ext string,
) (*Image, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	id := uuid.New()

	key := NewStorageKey(userID, id, ext)

	return &Image{
		id:         id,
		userID:     userID,
		storageKey: key,
		metadata:   meta,
		createdAt:  time.Now(),
	}, nil
}

func (i *Image) ID() uuid.UUID {
	return i.id
}

func (i *Image) UserID() uuid.UUID {
	return i.userID
}

func (i *Image) StorageKey() StorageKey {
	return i.storageKey
}

func (i *Image) Metadata() ImageMetadata {
	return i.metadata
}
