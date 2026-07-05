package transformation

type FilterParameters struct {
	Type      string            `json:"type"`
	Intensity float64           `json:"intensity,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
	Format    string            `json:"format,omitempty"`
}

func (p FilterParameters) Validate() error {
	if p.Type == "" {
		return ErrInvalidFilter
	}

	if p.Intensity < 0 || p.Intensity > 1 {
		return ErrInvalidIntensity
	}

	return nil
}
