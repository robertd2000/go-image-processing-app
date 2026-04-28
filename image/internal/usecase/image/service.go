package image

import (
	"context"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
)

type imageService struct {
	imageRepo         imageDomain.Repository
	storage           port.Storage
	metadataExtractor port.Extractor
}

func NewImageService(imageRepo imageDomain.Repository,
	storage port.Storage,
	metadataExtractor port.Extractor,
) *imageService {
	return &imageService{
		imageRepo:         imageRepo,
		storage:           storage,
		metadataExtractor: metadataExtractor,
	}
}

func (s *imageService) UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error) {
	return nil, nil
}
