package tokenpg

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/dberrors"
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
	var count int

	err := t.db.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM refresh_tokens
		WHERE user_id = $1
		AND token_hash = $2
		AND expires_at > NOW()
	`,
		userID,
		token,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("is token valid: %w", err)
	}

	return count > 0, nil
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
	if userID == uuid.Nil {
		return tokenDomain.ErrInvalidUserID
	}

	if token == "" {
		return tokenDomain.ErrInvalidToken
	}

	_, err := t.db.Exec(ctx, `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at,
			created_at
		) VALUES ($1, $2, $3, NOW)
	`, userID,
		token,
		expiresAt,
	)

	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return tokenDomain.ErrTokenAlreadyExists
		}

		return fmt.Errorf("save token: %w", err)
	}

	return nil
}

// Update implements token.TokenRepository.
func (t tokenRepository) Update(ctx context.Context, userID uuid.UUID, oldToken string, newToken string) error {
	panic("unimplemented")
}
