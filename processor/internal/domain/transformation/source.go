package transformation

type SourceImage struct {
	storageKey string
	mimeType   string

	width  int
	height int
}

func NewSourceImage(storageKey, mimeType string, width, height int) (SourceImage, error) {
	res := SourceImage{
		storageKey: storageKey,
		mimeType:   mimeType,
		width:      width,
		height:     height,
	}

	if err := res.Validate(); err != nil {
		return SourceImage{}, err
	}

	return res, nil
}

func (s SourceImage) Validate() error {
	if s.Height() < 0 {
		return ErrInvalidHeight
	}

	if s.Width() < 0 {
		return ErrInvalidWidth
	}

	return nil
}

func (s SourceImage) Width() int {
	return s.width
}
func (s SourceImage) Height() int {
	return s.height
}
func (s SourceImage) StorageKey() string {
	return s.storageKey
}
func (s SourceImage) MimeType() string {
	return s.mimeType
}
