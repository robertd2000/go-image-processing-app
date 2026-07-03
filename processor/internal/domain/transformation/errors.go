package transformation

import "errors"

var (
	ErrNotFound       = errors.New("transformation not found")
	ErrInvalidSpec    = errors.New("invalid transformation spec")
	ErrInvalidImageID = errors.New("invalid image id")
)
