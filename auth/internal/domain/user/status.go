package user

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBlocked  Status = "blocked"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive:
		return true
	default:
		return false
	}
}

func (s Status) IsActive() bool {
	return s == StatusActive
}

func (s Status) IsBlocked() bool {
	return s == StatusBlocked
}

func ParseStatus(v string) (Status, error) {
	s := Status(v)
	if !s.IsValid() {
		return "", ErrInvalidUserStatus
	}
	return s, nil
}
