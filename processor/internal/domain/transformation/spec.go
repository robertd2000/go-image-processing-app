package transformation

type TransformSpec struct {
	Operations []Operation `json:"operations"`
}

func (s TransformSpec) Validate() error {
	if len(s.Operations) == 0 {
		return ErrEmptyTransformation
	}

	for _, op := range s.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
	}

	return nil
}
