package image_test

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

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
	// outboxRepo port.OutboxRepository

	// tx port.TxManager
}

func (s *imageServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	// s.tx = &txmanagermem.FakeTxManager{}
	s.imageRepo = imagemem.NewInMemoryImageRepo()
	s.storage = storagemem.NewInMemoryStorage()
	s.metadataExtractor = imageInfra.NewMetadataExtractor()

	s.service = imageUsecase.NewImageService(s.imageRepo, s.storage, s.metadataExtractor)
}

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

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), output)
	assert.NotEqual(s.T(), uuid.Nil, output.ImageID)
}

func TestImageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(imageServiceTestSuite))
}

func generateTestImage() (*bytes.Buffer, int64) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)

	return buf, int64(buf.Len())
}
