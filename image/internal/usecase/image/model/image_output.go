package model

import (
	"time"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type ImageOutput struct {
	ImageID uuid.UUID
	UserID  uuid.UUID

	FileName string

	MimeType string
	Size     int64

	Width  int
	Height int

	URL string

	CreatedAt time.Time
}

func MapToImageOutput(
	img *imageDomain.Image,
	url string,
) *ImageOutput {
	return &ImageOutput{
		ImageID: img.ID(),
		UserID:  img.UserID(),

		FileName: img.OriginalName(),

		MimeType: img.Metadata().MimeType(),
		Size:     img.Metadata().Size(),

		Width:  img.Metadata().Width(),
		Height: img.Metadata().Height(),

		URL: url,

		CreatedAt: img.CreatedAt(),
	}
}
