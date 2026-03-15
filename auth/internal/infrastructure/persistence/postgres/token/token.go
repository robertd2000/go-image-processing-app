package tokenpg

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
)

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) tokenDomain.TokenRepository {
	return tokenRepository{
		db: db,
	}
}

// GetByToken implements token.TokenRepository.
func (t tokenRepository) GetByToken(ctx context.Context, token string) (*tokenDomain.Tokens, error) {
	panic("unimplemented")
}

// IsValid implements token.TokenRepository.
func (t tokenRepository) IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	panic("unimplemented")
}

// Revoke implements token.TokenRepository.
func (t tokenRepository) Revoke(ctx context.Context, userID uuid.UUID, token string) error {
	panic("unimplemented")
}

// RevokeByToken implements token.TokenRepository.
func (t tokenRepository) RevokeByToken(ctx context.Context, token string) error {
	panic("unimplemented")
}

// Save implements token.TokenRepository.
func (t tokenRepository) Save(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	panic("unimplemented")
}

// Update implements token.TokenRepository.
func (t tokenRepository) Update(ctx context.Context, userID uuid.UUID, oldToken string, newToken string) error {
	panic("unimplemented")
}
