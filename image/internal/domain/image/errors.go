package image

import "errors"

var (
	ErrInvalidImageDimensions      = errors.New("invalid dimensions")
	ErrInvalidImageSize            = errors.New("invalid size")
	ErrInvalidImageMissingMimeType = errors.New("mime type required")
	ErrInvalidImageMissingReader   = errors.New("reader required")
	ErrInvalidUserID               = errors.New("invalid user id")
	ErrExtractMetadata             = errors.New("invalid image metadata")
	ErrNotFound                    = errors.New("image not found")
	ErrAlreadyExists               = errors.New("image already exists")
	ErrInvalidPagination           = errors.New("invalid pagination params")
)
