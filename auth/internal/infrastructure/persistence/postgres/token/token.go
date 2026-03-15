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
	query := `
		SELECT * FROM refresh_tokens
		WHERE token_hash = $1
	`

	var tokenEntity *tokenDomain.Tokens

	err := t.db.QueryRow(ctx, query, token).Scan(tokenEntity)
	if err != nil {
		return nil, fmt.Errorf("get by token: %w", err)
	}

	return tokenEntity, nil
}

// IsValid implements token.TokenRepository.
func (t tokenRepository) IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	query := `
		SELECT COUNT(1)
		FROM refresh_tokens
		WHERE user_id = $1
		AND token_hash = $2
		AND expires_at > NOW()
	`
	var count int

	err := t.db.QueryRow(ctx, query, userID, token).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("is token valid: %w", err)
	}

	return count > 0, nil
}

// Revoke implements token.TokenRepository.
func (t tokenRepository) Revoke(ctx context.Context, userID uuid.UUID, token string) error {
	_, err := t.db.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE userID = $1
		AND token_hash = $2
	`,
		userID,
		token,
	)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}

// RevokeByToken implements token.TokenRepository.
func (t tokenRepository) RevokeByToken(ctx context.Context, token string) error {
	_, err := t.db.Exec(ctx, `
		DELETE FROM refresh_tokens
		AND token_hash = $1
		AND expires_at > NOW()
	`,
		token,
	)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
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
