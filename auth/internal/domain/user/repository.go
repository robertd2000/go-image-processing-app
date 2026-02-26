package user

import (
	"context"
	"os/user"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u user.User) (user.User, error)
	Update(ctx context.Context, id int64, user *user.User) error
	Delete(ctx context.Context, id int64) error
	FetchUserInfo(ctx context.Context, username string, password string) (user.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
}
