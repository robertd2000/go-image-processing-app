package port

import (
	"context"
	"io"
)

type Extractor interface {
	Extract(ctx context.Context, reader io.Reader) (ImageInfo, error)
}

type ImageInfo struct {
	Width    int
	Height   int
	MimeType string
}
