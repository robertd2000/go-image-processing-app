package model

import (
	"github.com/google/uuid"
)

type UserOutput struct {
	ID        uuid.UUID
	Username  string
	Email     string
	AvatarURL *string

	FirstName string
	LastName  string

	Bio      *string
	Location *string
	Website  *string

	IsPublic bool
	Theme    string
}
