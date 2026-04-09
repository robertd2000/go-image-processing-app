package port

import (
	"errors"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
)

type TokenGenerator interface {
	GenerateAccess(input model.ClaimsInput) (string, error)
	GenerateRefresh(userID uuid.UUID) (string, error)
	ValidateAccess(token string) (*auth.Claims, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)
