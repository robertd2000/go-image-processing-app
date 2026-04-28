package port

import (
	"context"
	"io"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type Extractor interface {
	Extract(ctx context.Context, reader io.Reader) (imageDomain.ImageMetadata, error)
}
