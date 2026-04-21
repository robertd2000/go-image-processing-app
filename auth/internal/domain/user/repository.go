package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type UserRepository interface {
	Create(ctx context.Context, tx port.Tx, u *AuthUser) error

	GetByEmail(ctx context.Context, email string) (*AuthUser, error)
	GetByUsername(ctx context.Context, username string) (*AuthUser, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuthUser, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	UpdateStatus(ctx context.Context, tx port.Tx, userID uuid.UUID, status Status) error
}
