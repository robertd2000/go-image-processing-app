package port

import (
	"io"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type Extractor interface {
	Extract(reader io.Reader) imageDomain.ImageMetadata
}
