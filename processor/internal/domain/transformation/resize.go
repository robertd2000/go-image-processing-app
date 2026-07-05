package transformation

type ResizeParameters struct {
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	MaintainRatio bool   `json:"maintain_ratio"`
	Quality       int    `json:"quality,omitempty"`
	Format        string `json:"format,omitempty"`
}

func (p ResizeParameters) Validate() error {
	if p.Width <= 0 {
		return ErrInvalidWidth
	}

	if p.Height <= 0 {
		return ErrInvalidHeight
	}

	if p.Quality != 0 && (p.Quality < 1 || p.Quality > 100) {
		return ErrInvalidQuality
	}

	return nil
}
