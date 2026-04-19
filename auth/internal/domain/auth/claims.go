// Package auth
package auth

import (
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID
	Email  string
	Roles  []string
}
