package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/postgres/dberrors"
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

func (s *userService) CreateFromEvent(ctx context.Context, input model.CreateUserInput) error {
	exists, err := s.userRepo.ExistsByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	username, err := userDomain.NewUsername(input.Username)
	if err != nil {
		return err
	}

	email, err := userDomain.NewEmail(input.Email)
	if err != nil {
		return err
	}

	user := userDomain.NewUser(
		input.ID,
		username,
		email,
	)

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if dberrors.IsUniqueViolation(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *userService) GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if user.Status() == userDomain.StatusInactive {
		return nil, userDomain.ErrUserNotFound
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

	if user.Status() == userDomain.StatusInactive {
		return nil, userDomain.ErrUserNotFound
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

func (s *userService) UpdateProfile(ctx context.Context, input model.UpdateProfileInput) error {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("find user: %w", err)
	}

	user.UpdateProfile(
		input.Bio,
		input.Location,
		input.Website,
		input.Birthday,
	)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return nil
}

func (s *userService) UpdateSettings(ctx context.Context, input model.UpdateSettingsInput) error {
	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("find user: %w", err)
	}

	if err := user.UpdateSettings(
		input.IsPublic,
		input.AllowNotifications,
		input.Theme,
	); err != nil {
		return fmt.Errorf("update user settings: %w", err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user settings: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("find user: %w", err)
	}

	if user.Status() == userDomain.StatusInactive {
		return userDomain.ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, userID)
}

func (s *userService) List(ctx context.Context, filter model.UserFilterInput) ([]*model.UserOutput, error) {
	userFilter, err := userDomain.NewUserFilter(filter.Limit, filter.Offset, nil, &filter.Search, filter.SortBy, filter.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("create user filter: %w", err)
	}

	users, err := s.userRepo.List(ctx, userFilter)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	var outputs []*model.UserOutput
	for _, user := range users {
		outputs = append(outputs, model.MapToOutput(user))
	}

	return outputs, nil
}

func (s *userService) Count(ctx context.Context, filter model.UserFilterInput) (int, error) {
	userFilter, err := userDomain.NewUserFilter(filter.Limit, filter.Offset, nil, &filter.Search, filter.SortBy, filter.SortOrder)
	if err != nil {
		return 0, fmt.Errorf("create user filter: %w", err)
	}

	count, err := s.userRepo.Count(ctx, userFilter)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}
