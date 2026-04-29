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
	roleDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/dberrors"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"go.uber.org/zap"
)

type userRepository struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewUserRepository(db *pgxpool.Pool, logger *zap.SugaredLogger) userDomain.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
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
			status,
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
			user.Status(),
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
			user.Status(),
			user.CreatedAt(),
		)
	}

	if err != nil {
		r.logger.Errorw("failed to create user",
			"user_id", user.ID(),
			"username", user.Username(),
			"email", user.Email(),
			"error", err,
		)

		if dberrors.IsUniqueViolation(err) {
			return userDomain.ErrUserAlreadyExists
		}

		return fmt.Errorf("userRepository.Create: %w", err)
	}

	r.logger.Infow("user created",
		"user_id", user.ID(),
		"username", user.Username(),
	)

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.created_at,
			COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM auth_users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.email = $1
		GROUP BY u.id
	`

	row := r.db.QueryRow(ctx, query, email)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Infow("user not found by email", "email", email)
			return nil, userDomain.ErrUserNotFound
		}

		r.logger.Errorw("failed to get user by email",
			"email", email,
			"error", err,
		)

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	r.logger.Debugw("user fetched by email",
		"user_id", user.ID(),
		"email", email,
	)

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.created_at,
			COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM auth_users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.username = $1
		GROUP BY u.id
	`

	row := r.db.QueryRow(ctx, query, username)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Infow("user not found by username", "username", username)
			return nil, userDomain.ErrUserNotFound
		}

		r.logger.Errorw("failed to get user by username",
			"username", username,
			"error", err,
		)

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	r.logger.Debugw("user fetched by username",
		"user_id", user.ID(),
		"username", username,
	)

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*userDomain.AuthUser, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password_hash,
			u.status,
			u.created_at,
			COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM auth_users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.id = $1
		GROUP BY u.id
	`

	row := r.db.QueryRow(ctx, query, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Infow("user not found by id", "user_id", userID)
			return nil, userDomain.ErrUserNotFound
		}

		r.logger.Errorw("failed to get user by id",
			"user_id", userID,
			"error", err,
		)

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	r.logger.Debugw("user fetched by id",
		"user_id", userID,
	)

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
		r.logger.Errorw("failed to check user exists by email",
			"email", email,
			"error", err,
		)
		return false, fmt.Errorf("check user exists by email: %w", err)
	}

	r.logger.Debugw("checked user exists by email",
		"email", email,
		"exists", exists,
	)

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
		r.logger.Errorw("failed to check user exists by username",
			"username", username,
			"error", err,
		)
		return false, fmt.Errorf("check user exists by email: %w", err)
	}

	r.logger.Debugw("checked user exists by username",
		"username", username,
		"exists", exists,
	)

	return exists, nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, tx port.Tx, userID uuid.UUID, status userDomain.Status) error {
	query := `
        UPDATE auth_users
        SET status = $1
        WHERE id = $2
    `

	err := tx.Exec(ctx, query, status, userID)
	if err != nil {
		r.logger.Errorw("failed to update user status",
			"user_id", userID,
			"status", status,
			"error", err,
		)
		return err
	}

	r.logger.Infow("user status updated",
		"user_id", userID,
		"status", status,
	)

	return nil
}

func scanUser(row pgx.Row) (*userDomain.AuthUser, error) {
	var (
		id           uuid.UUID
		username     string
		email        *string
		passwordHash string
		status       userDomain.Status
		createdAt    time.Time
		roleNames    []string
	)

	err := row.Scan(
		&id,
		&username,
		&email,
		&passwordHash,
		&status,
		&createdAt,
		&roleNames,
	)
	if err != nil {
		return nil, err
	}

	roles := make([]roleDomain.Role, 0, len(roleNames))

	for _, name := range roleNames {
		r, err := roleDomain.FromName(name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	return userDomain.NewUserFromDB(
		id,
		username,
		email,
		passwordHash,
		status,
		createdAt,
		roles,
	), nil
}
