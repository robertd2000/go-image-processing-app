package transformation

import "errors"

var (
	// Repository
	ErrNotFound = errors.New("transformation not found")

	// Validation
	ErrInvalidTransformationID = errors.New("invalid transformation id")
	ErrInvalidImageID          = errors.New("invalid image id")
	ErrInvalidStorageKey       = errors.New("invalid storage key")
	ErrInvalidMimeType         = errors.New("invalid mime type")
	ErrInvalidImageSize        = errors.New("invalid image size")
	ErrInvalidSpec             = errors.New("invalid transformation spec")

	// Domain
	ErrAlreadyExists           = errors.New("transformation already exists")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrAlreadyProcessing       = errors.New("transformation already processing")
	ErrAlreadyCompleted        = errors.New("transformation already completed")
)
