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
func (r tokenRepository) GetByHash(ctx context.Context, hash string) (*tokenDomain.Tokens, error) {
	query := `
		SELECT 
			id,
			user_id,
			token_hash,
			family_id,
			parent_id,
			expires_at,
			created_at,
			revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	row := r.db.QueryRow(ctx, query, hash)

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
func (r tokenRepository) IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
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

	err := r.db.QueryRow(ctx, query, userID, token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("is token valid: %w", err)
	}

	return exists, nil
}

func (t tokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`
	cmd, err := t.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("revoke token: rowsAffected is 0")
	}

	return nil
}

func (t tokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	if familyID == uuid.Nil {
		return nil // No family ID means nothing to revoke
	}

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE family_id = $1 AND revoked_at IS NULL
	`
	_, err := t.db.Exec(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("revoke token family: %w", err)
	}

	return nil
}

// Create implements token.TokenRepository.
func (r tokenRepository) Create(ctx context.Context, token *tokenDomain.Tokens, limit int) error {
	if token == nil {
		return tokenDomain.ErrInvalidToken
	}

	if token.UserID() == uuid.Nil {
		return tokenDomain.ErrInvalidUserID
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx,
		`
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			family_id,
			parent_id,
			created_at,
			expires_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`,
		token.ID(),
		token.UserID(),
		token.RefreshToken(),
		token.FamilyID(),
		token.ParentID(),
		token.CreatedAt(),
		token.ExpiresAt(),
	)
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return tokenDomain.ErrTokenAlreadyExists
		}
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE id IN (
			SELECT id FROM (
				SELECT id
				FROM refresh_tokens
				WHERE user_id = $1
				ORDER BY created_at DESC
				OFFSET $2
			) t
		)
	`, token.UserID(), limit)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Update implements token.TokenRepository.
func (r tokenRepository) Update(ctx context.Context, userID uuid.UUID, oldToken string, newToken string) error {
	query := `
		UPDATE refresh_tokens
		SET token_hash = $3
		WHERE user_id = $1 
		AND token_hash = $2
	`

	cmd, err := r.db.Exec(ctx, query, userID, oldToken, newToken)
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

func (r tokenRepository) Rotate(
	ctx context.Context,
	oldToken *tokenDomain.Tokens,
	newToken *tokenDomain.Tokens,
) (bool, error) {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Пытаемся "захватить" токен (revoke)
	cmd, err := tx.Exec(ctx, `
        UPDATE refresh_tokens
        SET revoked_at = NOW()
        WHERE id = $1 AND revoked_at IS NULL
    `, oldToken.ID())
	if err != nil {
		return false, fmt.Errorf("revoke old token: %w", err)
	}

	// ❗ КЛЮЧЕВОЙ МОМЕНТ
	if cmd.RowsAffected() == 0 {
		// токен уже был использован → REUSE ATTACK
		return true, tx.Commit(ctx) // commit чтобы зафиксировать "факт"
	}

	// 2. Вставляем новый токен
	_, err = tx.Exec(ctx, `
        INSERT INTO refresh_tokens (
            id,
            user_id,
            token_hash,
            family_id,
            parent_id,
            created_at,
            expires_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `,
		newToken.ID(),
		newToken.UserID(),
		newToken.RefreshToken(),
		newToken.FamilyID(),
		newToken.ParentID(),
		newToken.CreatedAt(),
		newToken.ExpiresAt(),
	)
	if err != nil {
		return false, fmt.Errorf("insert new token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}

	return false, nil
}

func scanToken(row pgx.Row) (*tokenDomain.Tokens, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		hash      string
		familyID  uuid.UUID
		parentID  uuid.UUID
		expiresAt time.Time
		createdAt time.Time
		revokedAt *time.Time
	)

	err := row.Scan(
		&id,
		&userID,
		&hash,
		&familyID,
		&parentID,
		&expiresAt,
		&createdAt,
		&revokedAt,
	)
	if err != nil {
		return nil, err
	}

	return tokenDomain.RestoreTokens(
		id,
		userID,
		hash,
		expiresAt,
		createdAt,
		revokedAt,
		familyID,
		parentID,
	), nil
}
