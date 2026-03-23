package tokenpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// GetByHash implements token.TokenRepository.
func (t tokenRepository) GetByHash(ctx context.Context, hash string) (*tokenDomain.Tokens, error) {
	query := `
		SELECT user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	row := t.db.QueryRow(ctx, query, hash)

	tokenEntity, err := scanToken(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tokenDomain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("get by token: %w", err)
	}

	return tokenEntity, nil
}

// IsValid implements token.TokenRepository.
func (t tokenRepository) IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM refresh_tokens
			WHERE user_id = $1
			AND token_hash = $2
			AND revoked_at IS NULL
			AND expires_at > NOW()
		)
	`
	var exists bool

	err := t.db.QueryRow(ctx, query, userID, token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("is token valid: %w", err)
	}

	return exists, nil
}

func (t tokenRepository) Revoke(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	cmd, err := t.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("revoke token: rowsAffected is 0")
	}

	return nil
}

// Create implements token.TokenRepository.
func (t tokenRepository) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	if userID == uuid.Nil {
		return tokenDomain.ErrInvalidUserID
	}

	if token == "" {
		return tokenDomain.ErrInvalidToken
	}

	query := `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at,
			created_at
		) VALUES ($1, $2, $3, NOW())
	`

	_, err := t.db.Exec(ctx, query, userID, token, expiresAt)
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
	query := `
		UPDATE refresh_tokens
		SET token_hash = $3
		WHERE user_id = $1 
		AND token_hash = $2
	`

	cmd, err := t.db.Exec(ctx, query, userID, oldToken, newToken)
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return tokenDomain.ErrTokenAlreadyExists
		}

		return fmt.Errorf("update token: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return tokenDomain.ErrTokenNotFound
	}

	return nil
}

func scanToken(row pgx.Row) (*tokenDomain.Tokens, error) {
	var (
		userID       uuid.UUID
		refreshToken string
		expiresAt    time.Time
		createdAt    time.Time
		revokedAt    *time.Time
	)

	err := row.Scan(
		&userID,
		&refreshToken,
		&expiresAt,
		&createdAt,
		&revokedAt,
	)
	if err != nil {
		return nil, err
	}

	return tokenDomain.NewTokens(userID, refreshToken, expiresAt)
}
