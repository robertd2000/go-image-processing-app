package user

import "time"

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
	isPublic bool,
	allowNotifications bool,
	theme string,
) {
	s.isPublic = isPublic
	s.allowNotifications = allowNotifications
	s.theme = theme
	s.updatedAt = time.Now()
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
