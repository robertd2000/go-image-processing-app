package user

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive:
		return true
	default:
		return false
	}
}

func ParseStatus(v string) (Status, error) {
	s := Status(v)
	if !s.IsValid() {
		return "", ErrInvalidUserStatus
	}
	return s, nil
}
