package model

import "github.com/google/uuid"

type ClaimsInput struct {
	UserID uuid.UUID
	Email  string
	Roles  []string
}
