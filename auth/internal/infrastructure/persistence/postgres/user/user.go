// Package userpg
package userpg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/dberrors"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) userDomain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(
	ctx context.Context,
	tx port.Tx,
	user *userDomain.AuthUser,
) error {
	query := `
		INSERT INTO auth_users (
			id,
			username,
			email,
			password_hash,
			enabled,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`

	var err error

	if tx != nil {
		err = tx.Exec(
			ctx,
			query,
			user.ID(),
			user.Username(),
			user.Email(),
			user.PasswordHash(),
			user.Enabled(),
			user.CreatedAt(),
		)
	} else {
		_, err = r.db.Exec(
			ctx,
			query,
			user.ID(),
			user.Username(),
			user.Email(),
			user.PasswordHash(),
			user.Enabled(),
			user.CreatedAt(),
		)
	}

	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return userDomain.ErrUserAlreadyExists
		}
		return fmt.Errorf("userRepository.Create: %w", err)
	}

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			enabled,
			created_at
		FROM auth_users
		WHERE email = $1
	`

	row := r.db.QueryRow(ctx, query, email)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			id,
			username,
			first_name,
			last_name,
			email,
			password_hash,
			enabled,
			created_at,
			modified_at,
			deleted_at
		FROM auth_users
		WHERE username = $1
	`

	row := r.db.QueryRow(ctx, query, username)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			id,
			username,
			first_name,
			last_name,
			email,
			password_hash,
			enabled,
			created_at,
			modified_at,
			deleted_at
		FROM auth_users
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userDomain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM auth_users
			WHERE email = $1
		)
	`

	var exists bool

	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user exists by email: %w", err)
	}

	return exists, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM auth_users
			WHERE username = $1
		)
	`

	var exists bool

	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user exists by email: %w", err)
	}

	return exists, nil
}

func (r *userRepository) Disable(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth_users
		SET enabled = false
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("disable user: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return userDomain.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Enable(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth_users
		SET enabled = true
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("enable user: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return userDomain.ErrUserNotFound
	}

	return nil
}

func scanUser(row pgx.Row) (*userDomain.AuthUser, error) {
	var (
		id           uuid.UUID
		username     string
		email        *string
		passwordHash string
		enabled      bool
		createdAt    time.Time
	)

	err := row.Scan(
		&id,
		&username,
		&email,
		&passwordHash,
		&enabled,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	return userDomain.NewUserFromDB(
		id,
		username,
		email,
		passwordHash,
		enabled,
		createdAt,
	), nil
}
