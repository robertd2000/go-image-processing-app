package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
)

type userSyncService struct {
	userRepo userDomain.UserRepository
}

func NewUserSyncService(userRepo userDomain.UserRepository) *userSyncService {
	return &userSyncService{
		userRepo: userRepo,
	}
}

func (s *userSyncService) Delete(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return userDomain.ErrInvalidUserID
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil
		}
		return err
	}

	if user.Status() == userDomain.StatusInactive {
		return nil
	}

	return s.userRepo.UpdateStatus(ctx, userID, userDomain.StatusInactive)
}
