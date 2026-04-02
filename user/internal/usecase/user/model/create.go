package model

import "github.com/google/uuid"

type CreateUserInput struct {
	ID       uuid.UUID
	Username string
	Email    string
}
