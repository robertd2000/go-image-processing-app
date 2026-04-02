package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

func (s *userService) GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return model.MapToOutput(user), nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.UserOutput, error) {
	userEmail, err := userDomain.NewEmail(email)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return model.MapToOutput(user), nil
}

func (s *userService) Update(ctx context.Context, input model.UpdateUserInput) error {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("find user: %w", err)
	}

	if input.Username != nil {
		username, err := userDomain.NewUsername(*input.Username)
		if err != nil {
			return err
		}

		exists, err := s.userRepo.ExistsByUsername(ctx, username)
		if err != nil {
			return fmt.Errorf("check username exists: %w", err)
		}
		if exists && user.Username() != username {
			return userDomain.ErrUsernameAlreadyExists
		}

		if err := user.ChangeUsername(username); err != nil {
			return err
		}
	}

	if input.Email != nil {
		email, err := userDomain.NewEmail(*input.Email)
		if err != nil {
			return err
		}

		exists, err := s.userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("check email exists: %w", err)
		}
		if exists && user.Email() != email {
			return userDomain.ErrEmailAlreadyExists
		}

		if err := user.ChangeEmail(email); err != nil {
			return err
		}
	}

	if input.FirstName != nil {
		if err := user.ChangeFirstName(*input.FirstName); err != nil {
			return err
		}
	}

	if input.LastName != nil {
		if err := user.ChangeLastname(*input.LastName); err != nil {
			return err
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *userService) Delete(id string) error {
	// TODO: implement delete user logic
	return nil
}
