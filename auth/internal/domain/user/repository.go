package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *AuthUser) error

	GetByEmail(ctx context.Context, email string) (*AuthUser, error)
	GetByUsername(ctx context.Context, username string) (*AuthUser, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuthUser, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	Disable(ctx context.Context, id uuid.UUID) error
	Enable(ctx context.Context, id uuid.UUID) error
}
