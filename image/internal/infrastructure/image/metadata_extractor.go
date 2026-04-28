package image

import (
	"context"
	"fmt"
	"image"
	"io"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type MetadataExtractor struct{}

func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

func (e *MetadataExtractor) Extract(ctx context.Context, reader io.Reader) (imageDomain.ImageMetadata, error) {
	cfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return imageDomain.ImageMetadata{}, fmt.Errorf("decode image config: %w", err)
	}

	mime := mapFormatToMime(format)

	meta, err := imageDomain.NewImageMetadata(
		cfg.Width,
		cfg.Height,
		0,
		mime,
	)
	if err != nil {
		return imageDomain.ImageMetadata{}, err
	}

	return meta, nil
}

func mapFormatToMime(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
