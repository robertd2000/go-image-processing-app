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
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	imageUsecase "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageService interface {
	UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error)
}

type imageServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service   ImageService
	imageRepo imageDomain.Repository
	storage   port.Storage

	metadataExtractor port.Extractor
}

func (s *imageServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.imageRepo = imagemem.NewInMemoryImageRepo()
	s.storage = storagemem.NewInMemoryStorage()
	s.metadataExtractor = imageInfra.NewMetadataExtractor()

	s.service = imageUsecase.NewImageService(s.imageRepo, s.storage, s.metadataExtractor)
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
