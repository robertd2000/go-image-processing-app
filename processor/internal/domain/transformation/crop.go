package transformation

type CropParameters struct {
	X int `json:"x"`
	Y int `json:"y"`

	Width  int `json:"width"`
	Height int `json:"height"`

	Format string `json:"format,omitempty"`
}

func (p CropParameters) Validate() error {
	if p.X < 0 {
		return ErrInvalidX
	}

	if p.Y < 0 {
		return ErrInvalidY
	}

	if p.Width <= 0 {
		return ErrInvalidWidth
	}

	if p.Height <= 0 {
		return ErrInvalidHeight
	}

	return nil
}
