package image

type ImageMetadata struct {
	Width    int
	Height   int
	Size     int64
	MimeType string
}

func NewImageMetadata(width, height int, size int64, mime string) (ImageMetadata, error) {
	if width <= 0 || height <= 0 {
		return ImageMetadata{}, ErrInvalidImageDimensions
	}
	if size <= 0 {
		return ImageMetadata{}, ErrInvalidImageSize
	}
	if mime == "" {
		return ImageMetadata{}, ErrInvalidImageMissingMimeType
	}

	return ImageMetadata{
		Width:    width,
		Height:   height,
		Size:     size,
		MimeType: mime,
	}, nil
}
