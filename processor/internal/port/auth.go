package port

import "github.com/google/uuid"

type AuthClaims struct {
	UserID uuid.UUID
	Roles  []string
}

type TokenValidator interface {
	ValidateAccess(tokenStr string) (*AuthClaims, error)
}
