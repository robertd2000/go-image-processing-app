// Package usermem
package usermem

import (
	"context"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
)

type userInMemoryRepository struct {
}

// ExistsByEmail implements user.UserRepository.
func (u *userInMemoryRepository) ExistsByEmail(ctx context.Context, email userDomain.Email) (bool, error) {
	panic("unimplemented")
}

// ExistsByUsername implements user.UserRepository.
func (u *userInMemoryRepository) ExistsByUsername(ctx context.Context, username userDomain.Username) (bool, error) {
	panic("unimplemented")
}

// Create implements user.UserRepository.
func (u *userInMemoryRepository) Create(ctx context.Context, user *userDomain.User) error {
	panic("unimplemented")
}

// Delete implements user.UserRepository.
func (u *userInMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByEmail implements user.UserRepository.
func (u *userInMemoryRepository) FindByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	panic("unimplemented")
}

// FindByID implements user.UserRepository.
func (u *userInMemoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	panic("unimplemented")
}

// Update implements user.UserRepository.
func (u *userInMemoryRepository) Update(ctx context.Context, user *userDomain.User) error {
	panic("unimplemented")
}

func NewUserRepository() userDomain.UserRepository {
	return &userInMemoryRepository{}
}
