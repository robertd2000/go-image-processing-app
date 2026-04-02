package model

import (
	"github.com/google/uuid"
)

type UserOutput struct {
	ID       uuid.UUID
	Username string
	Email    string

	Profile  UserProfileOutput
	Settings UserSettingsOutput
}

type UserProfileOutput struct {
	Bio      *string
	Location *string
	Website  *string
}

type UserSettingsOutput struct {
	IsPublic bool
	Theme    string
}
