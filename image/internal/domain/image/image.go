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
	id           uuid.UUID
	userID       uuid.UUID
	storageKey   StorageKey
	originalName string
	metadata     ImageMetadata
	createdAt    time.Time
	deletedAt    time.Time
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
		id:           id,
		userID:       userID,
		originalName: originalName,
		storageKey:   key,
		metadata:     meta,
		createdAt:    time.Now(),
	}, nil
}

func RestoreImage(
	id uuid.UUID,
	userID uuid.UUID,
	storageKey StorageKey,
	originalName string,
	meta ImageMetadata,
	createdAt time.Time,
	deletedAt time.Time,
) (*Image, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	return &Image{
		id:           id,
		userID:       userID,
		originalName: originalName,
		storageKey:   storageKey,
		metadata:     meta,
		createdAt:    createdAt,
		deletedAt:    deletedAt,
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

func (i *Image) OriginalName() string {
	return i.originalName
}

func (i *Image) Metadata() ImageMetadata {
	return i.metadata
}

func (i *Image) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Image) DeletedAt() time.Time {
	return i.deletedAt
}
