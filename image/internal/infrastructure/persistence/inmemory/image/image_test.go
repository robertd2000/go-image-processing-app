package imagemem_test

import (
	"context"
	"testing"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func newTestImage(userID uuid.UUID) *imageDomain.Image {
	meta, _ := imageDomain.NewImageMetadata(100, 100, 123, "image/png")

	img, _ := imageDomain.NewImage(
		userID,
		"test.png",
		meta,
		"png",
	)

	return img
}

func TestImageRepo_Save_And_GetByID(t *testing.T) {
	repo := imagemem.NewInMemoryImageRepo()

	userID := uuid.New()
	img := newTestImage(userID)

	err := repo.Save(context.Background(), img)
	assert.NoError(t, err)

	got, err := repo.GetByID(context.Background(), img.ID())
	assert.NoError(t, err)

	assert.Equal(t, img.ID(), got.ID())
	assert.Equal(t, img.UserID(), got.UserID())
	assert.Equal(t, img.StorageKey(), got.StorageKey())
	assert.Equal(t, img.Metadata(), got.Metadata())
}
