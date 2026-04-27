package image_test

import (
	"context"

	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
)

type ImageService interface {
	UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error)
}
