package image_test

import (
	"context"
	"testing"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
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

func TestImageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(imageServiceTestSuite))
}
