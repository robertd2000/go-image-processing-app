package transformation

type FormatParameters struct {
	TargetFormat string `json:"target_format"`

	Quality int `json:"quality,omitempty"`
}

func (p FormatParameters) Validate() error {
	if p.TargetFormat == "" {
		return ErrInvalidFormat
	}

	if p.Quality != 0 && (p.Quality < 1 || p.Quality > 100) {
		return ErrInvalidQuality
	}

	return nil
}
