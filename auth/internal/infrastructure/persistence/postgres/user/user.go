// Package userpg
package userpg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/dberrors"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) userDomain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *userDomain.User) error {
	query := `
		INSERT INTO users (
			id,
			username,
			first_name,
			last_name,
			email,
			password_hash,
			enabled,
			created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		user.ID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Email,
		user.PasswordHash,
		user.Enabled,
		user.CreatedAt,
	)
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return userDomain.ErrUserAlreadyExists
		}
		return fmt.Errorf("userRepository.Create: %w", err)
	}

	return nil
}

func (r *userRepository) Update(ctx context.Context, user *userDomain.User) error {
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userUD uuid.UUID) error {
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	return nil, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (r *userRepository) Disable(ctx context.Context, userUD uuid.UUID) error {
	return nil
}

func (r *userRepository) Enable(ctx context.Context, userUD uuid.UUID) error {
	return nil
}
