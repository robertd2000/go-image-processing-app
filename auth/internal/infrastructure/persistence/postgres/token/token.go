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
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"go.uber.org/zap"
)

type tokenRepository struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewTokenRepository(db *pgxpool.Pool, logger *zap.SugaredLogger) tokenDomain.TokenRepository {
	return tokenRepository{
		db:     db,
		logger: logger,
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
			r.logger.Infow("token not found by hash")
			return nil, tokenDomain.ErrTokenNotFound
		}

		r.logger.Errorw("failed to get token by hash",
			"error", err,
		)

		return nil, fmt.Errorf("get by token: %w", err)
	}

	r.logger.Debugw("token fetched by hash",
		"token_id", tokenEntity.ID(),
		"user_id", tokenEntity.UserID(),
	)

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
		r.logger.Errorw("failed to check token validity",
			"user_id", userID,
			"error", err,
		)
		return false, fmt.Errorf("is token valid: %w", err)
	}

	r.logger.Debugw("token validity checked",
		"user_id", userID,
		"is_valid", exists,
	)

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
		t.logger.Errorw("failed to revoke token",
			"token_id", tokenID,
			"error", err,
		)
		return fmt.Errorf("revoke token: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		t.logger.Warnw("token revoke affected 0 rows",
			"token_id", tokenID,
		)
		return fmt.Errorf("revoke token: rowsAffected is 0")
	}

	t.logger.Infow("token revoked",
		"token_id", tokenID,
	)

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
		t.logger.Errorw("failed to revoke token family",
			"family_id", familyID,
			"error", err,
		)
		return fmt.Errorf("revoke token family: %w", err)
	}

	t.logger.Infow("token family revoked",
		"family_id", familyID,
	)

	return nil
}

// Create implements token.TokenRepository.
func (r tokenRepository) Create(ctx context.Context, tx port.Tx, token *tokenDomain.Tokens, limit int) error {
	if token == nil {
		return tokenDomain.ErrInvalidToken
	}

	if token.UserID() == uuid.Nil {
		return tokenDomain.ErrInvalidUserID
	}

	// tx, err := r.db.Begin(ctx)
	// if err != nil {
	// 	r.logger.Errorw("failed to begin tx (create token)", "error", err)
	// 	return err
	// }
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err := tx.Exec(ctx,
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
		r.logger.Errorw("failed to insert token",
			"user_id", token.UserID(),
			"token_id", token.ID(),
			"error", err,
		)

		if dberrors.IsUniqueViolation(err) {
			return tokenDomain.ErrTokenAlreadyExists
		}
		return err
	}

	err = tx.Exec(ctx, `
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
		r.logger.Errorw("failed to cleanup old tokens",
			"user_id", token.UserID(),
			"limit", limit,
			"error", err,
		)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Errorw("failed to commit tx (create token)",
			"user_id", token.UserID(),
			"error", err,
		)
		return err
	}

	r.logger.Infow("token created",
		"user_id", token.UserID(),
		"token_id", token.ID(),
	)

	return nil
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
		r.logger.Errorw("failed to update token",
			"user_id", userID,
			"error", err,
		)

		if dberrors.IsUniqueViolation(err) {
			return tokenDomain.ErrTokenAlreadyExists
		}

		return fmt.Errorf("update token: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		r.logger.Warnw("token update affected 0 rows",
			"user_id", userID,
		)
		return tokenDomain.ErrTokenNotFound
	}

	r.logger.Infow("token updated",
		"user_id", userID,
	)

	return nil
}

func (r tokenRepository) Rotate(
	ctx context.Context,
	oldToken *tokenDomain.Tokens,
	newToken *tokenDomain.Tokens,
) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Errorw("failed to begin tx (rotate)", "error", err)
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, `
        UPDATE refresh_tokens
        SET revoked_at = NOW()
        WHERE id = $1 AND revoked_at IS NULL
    `, oldToken.ID())
	if err != nil {
		r.logger.Errorw("failed to revoke old token",
			"token_id", oldToken.ID(),
			"error", err,
		)
		return false, fmt.Errorf("revoke old token: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		r.logger.Warnw("token already revoked (possible replay attack)",
			"token_id", oldToken.ID(),
		)
		return true, tx.Commit(ctx)
	}

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
		r.logger.Errorw("failed to insert new token",
			"user_id", newToken.UserID(),
			"error", err,
		)
		return false, fmt.Errorf("insert new token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Errorw("failed to commit rotate tx",
			"user_id", newToken.UserID(),
			"error", err,
		)
		return false, fmt.Errorf("commit tx: %w", err)
	}

	r.logger.Infow("token rotated",
		"user_id", newToken.UserID(),
		"old_token_id", oldToken.ID(),
		"new_token_id", newToken.ID(),
	)

	return false, nil
}

func (r tokenRepository) DeleteByUserID(ctx context.Context, tx port.Tx, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()	
		WHERE user_id = $1
	`

	err := tx.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Errorw("failed to revoke tokens by user",
			"user_id", userID,
			"error", err,
		)
		return fmt.Errorf("delete token by user id: %w", err)
	}

	r.logger.Infow("all tokens revoked for user",
		"user_id", userID,
	)

	return nil
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
