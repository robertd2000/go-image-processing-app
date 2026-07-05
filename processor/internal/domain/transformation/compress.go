package transformation

type CompressParameters struct {
	Quality int `json:"quality"`

	Format string `json:"format"`
}

func (p CompressParameters) Validate() error {
	if p.Quality < 1 || p.Quality > 100 {
		return ErrInvalidQuality
	}

	if p.Format == "" {
		return ErrInvalidFormat
	}

	return nil
}
