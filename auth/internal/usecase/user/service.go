package user

import (
	"context"

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

func (s *userSyncService) UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return s.userRepo.UpdateStatus(ctx, userID, status)
}
