package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	ID       uuid.UUID
	Username string
	Email    string
}

type UpdateUserInput struct {
	UserID uuid.UUID

	Username  *string
	Email     *string
	FirstName *string
	LastName  *string
	AvatarURL *string
}

type UpdateProfileInput struct {
	UserID uuid.UUID

	Bio      *string
	Location *string
	Website  *string
	Birthday *time.Time
}

type UpdateSettingsInput struct {
	UserID uuid.UUID

	IsPublic           *bool
	AllowNotifications *bool
	Theme              *string
}
