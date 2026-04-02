package user

import (
	"errors"
	"time"
)

type UserSettings struct {
	isPublic           bool
	allowNotifications bool
	theme              string

	createdAt time.Time
	updatedAt time.Time
}

func NewSettings() *UserSettings {
	now := time.Now()

	return &UserSettings{
		isPublic:           true,
		allowNotifications: true,
		theme:              "light",
		createdAt:          now,
		updatedAt:          now,
	}
}

func (s *UserSettings) Update(
	isPublic, allowNotifications *bool,
	theme *string,
) error {
	if isPublic != nil {
		s.isPublic = *isPublic
	}

	if allowNotifications != nil {
		s.allowNotifications = *allowNotifications
	}

	if theme != nil {
		if err := validateTheme(*theme); err != nil {
			return err
		}
		s.theme = *theme
	}

	s.updatedAt = time.Now()
	return nil
}

func (s *UserSettings) IsPublic() bool {
	return s.isPublic
}

func (s *UserSettings) AllowNotifications() bool {
	return s.allowNotifications
}

func (s *UserSettings) Theme() string {
	return s.theme
}

func (s *UserSettings) CreatedAt() time.Time {
	return s.createdAt
}

func (s *UserSettings) UpdatedAt() time.Time {
	return s.updatedAt
}

func validateTheme(t string) error {
	switch t {
	case "light", "dark":
		return nil
	default:
		return errors.New("invalid theme")
	}
}
