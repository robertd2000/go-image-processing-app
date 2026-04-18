package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/user/internal/port"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindByUsername(ctx context.Context, username Username) (*User, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error

	ExistsByUsername(ctx context.Context, username Username) (bool, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	List(ctx context.Context, filter UserFilter) ([]*User, error)
	Count(ctx context.Context, filter UserFilter) (int, error)

	UpdateStatus(ctx context.Context, tx port.Tx, userID uuid.UUID, status UserStatus) error
}
