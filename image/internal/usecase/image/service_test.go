package image_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	imageInfra "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	storagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/storage"
	txmanagermem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/txmanager"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	imageUsecase "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageService interface {
	UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error)
	GetImage(ctx context.Context, imageID uuid.UUID) (*model.ImageOutput, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
}

type imageServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service   ImageService
	imageRepo imageDomain.Repository
	storage   port.Storage

	txManager port.TxManager

	metadataExtractor port.Extractor
}

func (s *imageServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.imageRepo = imagemem.NewInMemoryImageRepo()
	s.storage = storagemem.NewInMemoryStorage()
	s.metadataExtractor = imageInfra.NewMetadataExtractor()
	s.txManager = txmanagermem.NewFakeTxManager()

	s.service = imageUsecase.NewImageService(s.imageRepo, s.storage, s.metadataExtractor, s.txManager)
}

// SUCCESS

func (s *imageServiceTestSuite) TestUploadImage_Success() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := s.service.UploadImage(s.ctx, input)

	if !assert.NoError(s.T(), err) {
		return
	}

	assert.NotNil(s.T(), output)
	assert.NotEqual(s.T(), uuid.Nil, output.ImageID)
	assert.WithinDuration(s.T(), time.Now(), output.CreatedAt, time.Second)
}

// VALIDATION

func (s *imageServiceTestSuite) TestUploadImage_InvalidUserID() {
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   uuid.Nil,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidUserID)
}

func (s *imageServiceTestSuite) TestUploadImage_NoReader() {
	input := model.UploadImageInput{
		UserID: uuid.New(),
		Size:   100,
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidImageMissingReader)
}

func (s *imageServiceTestSuite) TestUploadImage_InvalidSize() {
	buf, _ := generateTestImage()

	input := model.UploadImageInput{
		UserID: uuid.New(),
		Reader: bytes.NewReader(buf.Bytes()),
		Size:   0,
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidImageSize)
}

// INVALID IMAGE

func (s *imageServiceTestSuite) TestUploadImage_InvalidImageData() {
	input := model.UploadImageInput{
		UserID: uuid.New(),
		Reader: bytes.NewReader([]byte("not-an-image")),
		Size:   int64(len("not-an-image")),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
}

// STORAGE ERROR

type failingStorage struct {
	port.Storage
}

func (f *failingStorage) Put(ctx context.Context, key string, r io.Reader, size int64, mime string) error {
	return errors.New("storage error")
}

func (s *imageServiceTestSuite) TestUploadImage_StorageFails() {
	userID := uuid.New()
	buf, size := generateTestImage()

	s.service = imageUsecase.NewImageService(
		s.imageRepo,
		&failingStorage{},
		s.metadataExtractor,
		s.txManager,
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "storage")
}

// EXTENSION FALLBACK

func (s *imageServiceTestSuite) TestUploadImage_UnknownMime_UsesFilenameExt() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.unknownext",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.NoError(s.T(), err)
}

// LARGE FILE LIMIT
func (s *imageServiceTestSuite) TestUploadImage_TooLarge() {
	userID := uuid.New()

	large := bytes.Repeat([]byte("a"), 20<<20) // 20MB

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "big.png",
		Size:     int64(len(large)),
		Reader:   bytes.NewReader(large),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
}

// REPO ERROR + ROLLBACK

type failingRepo struct {
	imageDomain.Repository
}

func (f *failingRepo) Save(ctx context.Context, tx port.Tx, img *imageDomain.Image) error {
	return errors.New("repo error")
}

func (s *imageServiceTestSuite) TestUploadImage_RepoFails_ShouldRollbackStorage() {
	userID := uuid.New()
	buf, size := generateTestImage()

	spy := storagemem.NewSpyStorage()

	s.service = imageUsecase.NewImageService(
		&failingRepo{},
		spy,
		s.metadataExtractor,
		s.txManager,
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)

	assert.True(s.T(), spy.PutCalled)
	assert.True(s.T(), spy.DeleteCalled)
}

// GetImage tests

func (s *imageServiceTestSuite) TestGetImage_Success() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)
	s.Require().NotNil(uploadRes)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Equal(uploadRes.ImageID, res.ImageID)
	s.NotEmpty(res.URL)
	s.Equal("test.png", res.FileName)
	s.Equal(int64(size), res.Size)
	s.Equal("image/png", res.MimeType)
}

func (s *imageServiceTestSuite) TestGetImage_NotFound() {
	res, err := s.service.GetImage(s.ctx, uuid.New())

	s.Require().Error(err)
	s.Nil(res)
}

func (s *imageServiceTestSuite) TestGetImage_StorageError() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	err = s.storage.Delete(s.ctx, string(img.StorageKey()))
	s.Require().NoError(err)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)

	s.Require().Error(err)
	s.Nil(res)
}

func (s *imageServiceTestSuite) TestGetImage_MetadataMapping() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	s.Equal(10, res.Width)
	s.Equal(10, res.Height)
}

// DeleteImage

func (s *imageServiceTestSuite) TestDeleteImage_Success() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().NoError(err)

	_, err = s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	s.Require().Error(err)

	img, err := s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	if err == nil {
		err = s.storage.Delete(s.ctx, string(img.StorageKey()))
		s.Require().Error(err)
	}
}

func (s *imageServiceTestSuite) TestDeleteImage_NotFound() {
	err := s.service.DeleteImage(s.ctx, uuid.New())

	s.Require().Error(err)
}

// HELPERS

func generateTestImage() (*bytes.Buffer, int64) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)

	return buf, int64(buf.Len())
}

func TestImageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(imageServiceTestSuite))
}
