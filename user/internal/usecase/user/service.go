package user

import (
	"context"

	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
)

type userService struct {
	userRepo userDomain.UserRepository
}

func NewUserService(userRepo userDomain.UserRepository) *userService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Create(ctx context.Context, input model.CreateUserInput) error {
	username, err := userDomain.NewUsername(input.Username)
	if err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return userDomain.ErrUsernameAlreadyExists
	}

	email, err := userDomain.NewEmail(input.Email)
	if err != nil {
		return err
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return userDomain.ErrEmailAlreadyExists
	}

	user := userDomain.NewUser(
		input.ID,
		username,
		email,
	)

	return s.userRepo.Create(ctx, user)
}

func (s *userService) GetByID(id string) (*userDomain.User, error) {
	// TODO: implement get user by ID logic
	return nil, nil
}

func (s *userService) Update(user *userDomain.User) error {
	// TODO: implement update user logic
	return nil
}

func (s *userService) Delete(id string) error {
	// TODO: implement delete user logic
	return nil
}
