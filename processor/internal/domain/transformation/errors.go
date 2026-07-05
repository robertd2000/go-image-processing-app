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
	ErrInvalidResultKey        = errors.New("invalid result key")
	ErrInvalidErrorMessage     = errors.New("invalid error message")
)

var (
	ErrEmptyTransformation = errors.New("transformation contains no operations")

	ErrOperationIsEmpty   = errors.New("operation is empty")
	ErrMultipleOperations = errors.New("operation contains multiple transformations")

	ErrInvalidWidth  = errors.New("invalid width")
	ErrInvalidHeight = errors.New("invalid height")

	ErrInvalidX = errors.New("invalid x")
	ErrInvalidY = errors.New("invalid y")

	ErrInvalidQuality = errors.New("invalid quality")

	ErrInvalidAngle = errors.New("invalid angle")

	ErrInvalidFilter    = errors.New("invalid filter")
	ErrInvalidIntensity = errors.New("invalid intensity")

	ErrInvalidWatermark = errors.New("invalid watermark")
	ErrInvalidOpacity   = errors.New("invalid opacity")
	ErrInvalidScale     = errors.New("invalid scale")
	ErrInvalidPosition  = errors.New("invalid position")

	ErrInvalidFormat = errors.New("invalid format")

	ErrInvalidFileSize = errors.New("invalid file size")
)
