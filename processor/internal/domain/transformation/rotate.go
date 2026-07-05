package transformation

type RotateParameters struct {
	Angle  float64 `json:"angle"`
	Format string  `json:"format,omitempty"`
}

func (p RotateParameters) Validate() error {
	if p.Angle < -360 || p.Angle > 360 {
		return ErrInvalidAngle
	}

	return nil
}
