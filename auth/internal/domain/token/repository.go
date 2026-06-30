// Package token
package token

import (
	"context"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

type TokenRepository interface {
	Create(ctx context.Context, tx txtx.Tx, token *Tokens, limit int) error
	Rotate(ctx context.Context, tx txtx.Tx, oldToken *Tokens, newToken *Tokens) (bool, error)
	Revoke(ctx context.Context, tx txtx.Tx, tokenID uuid.UUID) error
	RevokeFamily(ctx context.Context, tx txtx.Tx, familyID uuid.UUID) error
	Update(ctx context.Context, tx txtx.Tx, userID uuid.UUID, oldToken, newToken string) error
	DeleteByUserID(ctx context.Context, tx txtx.Tx, userID uuid.UUID) error
	GetByHash(ctx context.Context, hash string) (*Tokens, error)
	IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error)
}

type TokenGenerator interface {
	Generate(userID uuid.UUID, email string) (string, error)
	Validate(toke string) (uuid.UUID, error)
	GenerateAccess(userID uuid.UUID) (string, error)
	GenerateRefresh(userID uuid.UUID) (string, error)
	ValidateAccess(token string) (uuid.UUID, error)
	ValidateRefresh(token string) (uuid.UUID, error)
}
