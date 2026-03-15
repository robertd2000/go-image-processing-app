// Package token
package token

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenRepository interface {
	Save(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	Update(ctx context.Context, userID uuid.UUID, oldToken, newToken string) error
	Revoke(ctx context.Context, userID uuid.UUID, token string) error
	RevokeByToken(ctx context.Context, token string) error
	GetByToken(ctx context.Context, token string) (*Tokens, error)
}

type TokenGenerator interface {
	Generate(userID uuid.UUID, email string) (string, error)
	Validate(toke string) (uuid.UUID, error)
	GenerateAccess(userID uuid.UUID) (string, error)
	GenerateRefresh(userID uuid.UUID) (string, error)
	ValidateAccess(token string) (uuid.UUID, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}
