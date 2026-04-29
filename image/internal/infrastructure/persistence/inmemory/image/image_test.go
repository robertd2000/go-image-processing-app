package imagemem_test

import (
	"context"
	"sync"
	"testing"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	txmanagermem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/txmanager"
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

	err := s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img)
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

func (s *ImageRepoSuite) TestSave_Duplicate() {
	userID := uuid.New()
	img := s.newImage(userID)

	err := s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img)
	assert.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img)
	assert.ErrorIs(s.T(), err, imageDomain.ErrAlreadyExists)
}

func (s *ImageRepoSuite) TestGetByUser_Basic() {
	user1 := uuid.New()
	user2 := uuid.New()

	img1 := s.newImage(user1)
	img2 := s.newImage(user1)
	img3 := s.newImage(user2)

	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img1)
	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img2)
	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img3)

	res, err := s.repo.GetByUser(s.ctx, user1, 10, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res, 2)

	for _, img := range res {
		assert.Equal(s.T(), user1, img.UserID())
	}
}

func (s *ImageRepoSuite) TestGetByUser_Pagination() {
	user := uuid.New()

	for range 10 {
		_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, s.newImage(user))
	}

	res, err := s.repo.GetByUser(s.ctx, user, 5, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res, 5)

	res2, err := s.repo.GetByUser(s.ctx, user, 5, 5)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res2, 5)
}

func (s *ImageRepoSuite) TestGetByUser_Pagination_EdgeCases() {
	user := uuid.New()

	for range 3 {
		_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, s.newImage(user))
	}

	res, err := s.repo.GetByUser(s.ctx, user, 10, 100)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res, 0)

	res, err = s.repo.GetByUser(s.ctx, user, 0, 0)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), res, 0)
}

func (s *ImageRepoSuite) TestConcurrent_SaveAndGet() {
	user := uuid.New()

	const workers = 100

	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {

			img := s.newImage(user)

			err := s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img)
			assert.NoError(s.T(), err)

			got, err := s.repo.GetByID(s.ctx, img.ID())
			assert.NoError(s.T(), err)

			assert.Equal(s.T(), img.ID(), got.ID())
		})
	}

	wg.Wait()
}

func (s *ImageRepoSuite) TestConcurrent_GetByUser() {
	user := uuid.New()

	for range 50 {
		_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, s.newImage(user))
	}

	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {

			res, err := s.repo.GetByUser(s.ctx, user, 10, 0)
			assert.NoError(s.T(), err)
			assert.LessOrEqual(s.T(), len(res), 10)
		})
	}

	wg.Wait()
}

func (s *ImageRepoSuite) TestDelete_HidesFromGetByID() {
	userID := uuid.New()
	img := s.newImage(userID)

	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img)

	err := s.repo.Delete(s.ctx, img.ID())
	s.Require().NoError(err)

	_, err = s.repo.GetByID(s.ctx, img.ID())
	s.ErrorIs(err, imageDomain.ErrNotFound)
}

func (s *ImageRepoSuite) TestDelete_HidesFromGetByUser() {
	userID := uuid.New()

	img1 := s.newImage(userID)
	img2 := s.newImage(userID)

	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img1)
	_ = s.repo.Save(s.ctx, &txmanagermem.FakeTx{}, img2)

	_ = s.repo.Delete(s.ctx, img1.ID())

	res, err := s.repo.GetByUser(s.ctx, userID, 10, 0)
	s.Require().NoError(err)

	s.Len(res, 1)
	s.Equal(img2.ID(), res[0].ID())
}

func TestImageRepoSuite(t *testing.T) {
	suite.Run(t, new(ImageRepoSuite))
}
