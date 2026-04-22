package user

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBanned   Status = "banned"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusBanned:
		return true
	default:
		return false
	}
}

func (s Status) IsActive() bool {
	return s == StatusActive
}

func (s Status) IsBanned() bool {
	return s == StatusBanned
}

func ParseStatus(v string) (Status, error) {
	s := Status(v)
	if !s.IsValid() {
		return "", ErrInvalidUserStatus
	}
	return s, nil
}
