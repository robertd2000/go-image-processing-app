package port

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/auth"
)

type ClaimsInput struct {
	UserID uuid.UUID
	Email  string
	Roles  []string
}

type TokenGenerator interface {
	GenerateAccess(input ClaimsInput) (string, error)
	GenerateRefresh(userID uuid.UUID, ttl time.Duration) (string, error)
	ValidateAccess(token string) (*auth.Claims, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)
