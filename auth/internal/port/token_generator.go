package port

import (
	"errors"

	"github.com/google/uuid"
)

type TokenGenerator interface {
	Generate(userID uuid.UUID, email string) (string, error)
	Validate(toke string) (uuid.UUID, error)
	GenerateAccess(userID uuid.UUID) (string, error)
	GenerateRefresh(userID uuid.UUID) (string, error)
	ValidateAccess(token string) (uuid.UUID, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)
