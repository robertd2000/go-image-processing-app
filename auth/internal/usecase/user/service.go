package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type userSyncService struct {
	userRepo  userDomain.UserRepository
	tokenRepo tokenDomain.TokenRepository
	txManager port.TxManager
}

func NewUserSyncService(txManager port.TxManager, userRepo userDomain.UserRepository, tokenRepo tokenDomain.TokenRepository) *userSyncService {
	return &userSyncService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		txManager: txManager,
	}
}

func (s *userSyncService) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {
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

		if err := s.userRepo.UpdateStatus(ctx, tx, userID, userDomain.StatusInactive); err != nil {
			return err
		}

		if err := s.tokenRepo.DeleteByUserID(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *userSyncService) Ban(ctx context.Context, userID uuid.UUID) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {
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

		if user.Status() != userDomain.StatusActive {
			return nil
		}

		if err := s.userRepo.UpdateStatus(ctx, tx, userID, userDomain.StatusBanned); err != nil {
			return err
		}

		if err := s.tokenRepo.DeleteByUserID(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}
