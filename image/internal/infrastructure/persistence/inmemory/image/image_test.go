package imagemem_test

import (
	"context"
	"testing"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/google/uuid"
)

type ImageRepoSuite struct {
	suite.Suite

	repo imageDomain.Repository
	ctx  context.Context
}

func (s *ImageRepoSuite) SetupTest() {
	s.repo = imagemem.NewInMemoryImageRepo()
	s.ctx = context.Background()
}

func (s *ImageRepoSuite) newImage(userID uuid.UUID) *imageDomain.Image {
	meta, _ := imageDomain.NewImageMetadata(100, 100, 123, "image/png")

	img, _ := imageDomain.NewImage(
		userID,
		"test.png",
		meta,
		"png",
	)

	return img
}

func (s *ImageRepoSuite) TestSaveAndGetByID() {
	userID := uuid.New()
	img := s.newImage(userID)

	err := s.repo.Save(s.ctx, img)
	assert.NoError(s.T(), err)

	got, err := s.repo.GetByID(s.ctx, img.ID())
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), img.ID(), got.ID())
	assert.Equal(s.T(), img.UserID(), got.UserID())
	assert.Equal(s.T(), img.StorageKey(), got.StorageKey())
	assert.Equal(s.T(), img.Metadata(), got.Metadata())
}

func (s *ImageRepoSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(s.ctx, uuid.New())
	assert.ErrorIs(s.T(), err, imageDomain.ErrNotFound)
}

func TestImageRepoSuite(t *testing.T) {
	suite.Run(t, new(ImageRepoSuite))
}
