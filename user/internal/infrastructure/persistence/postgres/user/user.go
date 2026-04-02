// Package userpg
package userpg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/postgres/dberrors"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *userDomain.User) error {
	query := `INSERT INTO users (id, username, email) VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, query, user.ID(), user.Username(), user.Email())
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return userDomain.ErrUserAlreadyExists
		}
		return fmt.Errorf("userRepository.Create: %w", err)
	}

	return nil
}

func (r *userRepository) FindByID(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	// Implementation of the FindByID method
	return nil, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email userDomain.Email) (*userDomain.User, error) {
	// Implementation of the FindByEmail method
	return nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *userDomain.User) error {
	// Implementation of the Update method
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	// Implementation of the Delete method
	return nil
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username userDomain.Username) (bool, error) {
	// Implementation of the ExistsByUsername method
	return false, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email userDomain.Email) (bool, error) {
	// Implementation of the ExistsByEmail method
	return false, nil
}
