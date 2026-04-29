// Package token
package token

import (
	"context"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type TokenRepository interface {
	Create(ctx context.Context, tx port.Tx, token *Tokens, limit int) error
	Update(ctx context.Context, userID uuid.UUID, oldToken, newToken string) error
	Revoke(ctx context.Context, tokenID uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	Rotate(ctx context.Context,
		oldToken *Tokens,
		newToken *Tokens) (bool, error)
	GetByHash(ctx context.Context, token string) (*Tokens, error)
	DeleteByUserID(ctx context.Context, tx port.Tx, userID uuid.UUID) error
}

type TokenGenerator interface {
	Generate(userID uuid.UUID, email string) (string, error)
	Validate(toke string) (uuid.UUID, error)
	GenerateAccess(userID uuid.UUID) (string, error)
	GenerateRefresh(userID uuid.UUID) (string, error)
	ValidateAccess(token string) (uuid.UUID, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}
