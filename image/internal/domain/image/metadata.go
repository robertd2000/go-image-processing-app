package image

type ImageMetadata struct {
	width    int
	height   int
	size     int64
	mimeType string
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
		width:    width,
		height:   height,
		size:     size,
		mimeType: mime,
	}, nil
}

func (i ImageMetadata) Width() int {
	return i.width
}

func (i ImageMetadata) Height() int {
	return i.height
}

func (i ImageMetadata) Size() int64 {
	return i.size
}

func (i ImageMetadata) MimeType() string {
	return i.mimeType
}
