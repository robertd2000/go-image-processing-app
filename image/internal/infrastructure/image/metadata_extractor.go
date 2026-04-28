package image

import (
	"context"
	"fmt"
	"image"
	"io"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type MetadataExtractor struct{}

func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

func (e *MetadataExtractor) Extract(ctx context.Context, reader io.Reader) (port.ImageInfo, error) {
	cfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return port.ImageInfo{}, fmt.Errorf("decode image config: %w", err)
	}

	return port.ImageInfo{
		Width:    cfg.Width,
		Height:   cfg.Height,
		MimeType: mapFormatToMime(format),
	}, nil
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
