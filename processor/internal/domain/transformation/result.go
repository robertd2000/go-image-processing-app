package transformation

type ResultImage struct {
	storageKey string
	mimeType   string

	width  int
	height int

	size int64
}

func NewResultImage(
	storageKey string,
	mimeType string,
	width, height int,
	size int64,
) (ResultImage, error) {
	r := ResultImage{
		storageKey: storageKey,
		mimeType:   mimeType,
		width:      width,
		height:     height,
		size:       size,
	}

	if err := r.Validate(); err != nil {
		return ResultImage{}, err
	}

	return r, nil
}

func (r ResultImage) Validate() error {
	if r.storageKey == "" {
		return ErrInvalidStorageKey
	}

	if r.mimeType == "" {
		return ErrInvalidMimeType
	}

	if r.width <= 0 {
		return ErrInvalidWidth
	}

	if r.height <= 0 {
		return ErrInvalidHeight
	}

	if r.size <= 0 {
		return ErrInvalidFileSize
	}

	return nil
}

func (r ResultImage) StorageKey() string {
	return r.storageKey
}

func (r ResultImage) MimeType() string {
	return r.mimeType
}

func (r ResultImage) Width() int {
	return r.width
}

func (r ResultImage) Height() int {
	return r.height
}

func (r ResultImage) Size() int64 {
	return r.size
}
