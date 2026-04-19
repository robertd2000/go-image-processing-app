// Package jwt
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/auth"
)

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
	TokenGeneric TokenType = "generic"
)

func (t TokenType) String() string {
	return string(t)
}

type CustomClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func (c CustomClaims) VerifyIssuer(s string, true bool) bool {
	return c.Issuer == s
}

func toJWTClaims(c auth.Claims) CustomClaims {
	return CustomClaims{
		UserID: c.UserID,
		Roles:  c.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func toDomainClaims(c CustomClaims) *auth.Claims {
	return &auth.Claims{
		UserID: c.UserID,
		Roles:  c.Roles,
	}
}
