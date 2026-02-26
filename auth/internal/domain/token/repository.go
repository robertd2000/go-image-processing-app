// Package token
package token

import (
	"context"

	"github.com/google/uuid"
)

type TokenRepository interface {
	Save(ctx context.Context, userID uuid.UUID, token string) error
	IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	Update(ctx context.Context, userID uuid.UUID, oldToken, newToken string) error
	Revoke(ctx context.Context, userID uuid.UUID, token string) error
	RevokeByToken(ctx context.Context, token string) error
	GetByToken(ctx context.Context, token string) (*Tokens, error)
}
