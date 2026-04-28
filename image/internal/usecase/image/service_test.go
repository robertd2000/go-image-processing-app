package image_test

import (
	"context"
	"testing"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
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
	// outboxRepo port.OutboxRepository

	// tx port.TxManager
}

func (s *imageServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	// s.imageRepo = usermem.NewUserRepository()
	// s.outboxRepo = outboxmem.NewRepository()

	// s.tx = &txmanagermem.FakeTxManager{}

	// s.service =
}

// func (s *imageServiceTestSuite) TestUploadImage_Success() {
// 	userID := uuid.New()
// 	filename := "test1"
// 	size := 10000
// 	input := model.UploadImageInput{
// 		UserID:   userID,
// 		Filename: filename,
// 		Size:     int64(size),
// 	}

// 	output, err := s.service.UploadImage(s.ctx, input)
// 	assert.NoError(s.T(), err)
// 	assert.NotNil(s.T(), output)
// }

func TestImageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(imageServiceTestSuite))
}
