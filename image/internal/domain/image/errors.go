package image

import "errors"

var (
	ErrInvalidImageDimensions      = errors.New("invalid dimensions")
	ErrInvalidImageSize            = errors.New("invalid size")
	ErrInvalidImageMissingMimeType = errors.New("mime type required")
	ErrInvalidUserID               = errors.New("invalid user id")
	ErrExtractMetadata             = errors.New("invalid image metadata")
)
