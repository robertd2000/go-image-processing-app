package transformation

type WatermarkParameters struct {
	StorageKey string `json:"storage_key"`

	Position string `json:"position"`

	Opacity float64 `json:"opacity"`

	Scale float64 `json:"scale"`

	Format string `json:"format,omitempty"`
}

func (p WatermarkParameters) Validate() error {
	if p.StorageKey == "" {
		return ErrInvalidWatermark
	}

	if p.Position == "" {
		return ErrInvalidPosition
	}

	if p.Opacity < 0 || p.Opacity > 1 {
		return ErrInvalidOpacity
	}

	if p.Scale <= 0 {
		return ErrInvalidScale
	}

	return nil
}
