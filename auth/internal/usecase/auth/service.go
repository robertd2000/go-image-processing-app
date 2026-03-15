// Package auth
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/validation"
)

type authService struct {
	userRepo    userDomain.UserRepository
	refreshRepo tokensDomain.TokenRepository
	hasher      port.PasswordHasher
	tokenGen    port.TokenGenerator
}

func NewAuthService(userRepo userDomain.UserRepository, refreshRepo tokensDomain.TokenRepository, hasher port.PasswordHasher, tokenGen port.TokenGenerator) *authService {
	return &authService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		hasher:      hasher,
		tokenGen:    tokenGen,
	}
}

func (s *authService) Register(ctx context.Context, username, firstname, lastname, email, password string) error {
	if err := validation.ValidateEmail(email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return err
	}

	if err := validation.ValidateUsername(username); err != nil {
		return err
	}
	//
	// exists, err := s.userRepo.ExistsByEmail(ctx, email)
	// if err != nil {
	// 	return fmt.Errorf("find user by email: %w", err)
	// }
	// if exists {
	// 	return userDomain.ErrUserAlreadyExists
	// }
	//
	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := userDomain.CreateUser(username, firstname, lastname, &email, hashed)
	if err != nil {
		return err
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserAlreadyExists) {
			return err
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, email string, password string) (*tokensDomain.Tokens, error) {
	if err := validation.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil, userDomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !user.Enabled() {
		return nil, userDomain.ErrUserDisabled
	}

	if !s.hasher.Compare(user.PasswordHash(), password) {
		return nil, userDomain.ErrWrongCreadentials
	}

	return s.generateTokens(ctx, user.ID())
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*tokensDomain.Tokens, error) {
	userID, err := s.tokenGen.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, err
	}

	valid, err := s.refreshRepo.IsValid(ctx, userID, refreshToken)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, tokensDomain.ErrInvalidToken
	}

	if err := s.refreshRepo.Revoke(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, userID)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshRepo.RevokeByToken(ctx, refreshToken)
}

func (s *authService) generateTokens(ctx context.Context, userID uuid.UUID) (*tokensDomain.Tokens, error) {
	access, err := s.tokenGen.GenerateAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.tokenGen.GenerateRefresh(userID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.refreshRepo.Save(ctx, userID, refresh); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	tokens, err := tokensDomain.NewTokens(userID, access, refresh, time.Now().Add(time.Hour))
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
