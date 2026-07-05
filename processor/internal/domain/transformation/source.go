package transformation

type SourceImage struct {
	StorageKey string
	MimeType   string

	Width  int
	Height int
}

func NewSourceImage(storageKey, mimeType string, width, height int) (SourceImage, error) {
	res := SourceImage{
		StorageKey: storageKey,
		MimeType:   mimeType,
		Width:      width,
		Height:     height,
	}

	if err := res.Validate(); err != nil {
		return SourceImage{}, err
	}

	return res, nil
}

func (s SourceImage) Validate() error {
	if s.Height < 0 {
		return ErrInvalidHeight
	}

	if s.Width < 0 {
		return ErrInvalidWidth
	}

	return nil
}
