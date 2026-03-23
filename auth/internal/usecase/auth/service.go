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
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/dto"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/validation"
)

var sessionLimit = 5

type authService struct {
	userRepo    userDomain.UserRepository
	refreshRepo tokensDomain.TokenRepository

	tokenGen       port.TokenGenerator
	passwordHasher port.PasswordHasher
	tokenHasher    port.TokenHasher

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	userRepo userDomain.UserRepository,
	refreshRepo tokensDomain.TokenRepository,
	passwordHasher port.PasswordHasher,
	tokenHasher port.TokenHasher,
	tokenGen port.TokenGenerator,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *authService {
	return &authService{
		userRepo:       userRepo,
		refreshRepo:    refreshRepo,
		tokenGen:       tokenGen,
		passwordHasher: passwordHasher,
		tokenHasher:    tokenHasher,
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
	hashed, err := s.passwordHasher.Hash(password)
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

func (s *authService) Login(ctx context.Context, email string, password string) (*dto.TokenPair, error) {
	if err := validation.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, userDomain.ErrWrongCreadentials
	}

	if !user.Enabled() {
		return nil, userDomain.ErrUserDisabled
	}

	if !s.passwordHasher.Compare(user.PasswordHash(), password) {
		return nil, userDomain.ErrWrongCreadentials
	}

	return s.generateTokens(ctx, user.ID())
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
	now := time.Now()

	hash := s.tokenHasher.Hash(refreshToken)
	token, err := s.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, tokensDomain.ErrInvalidToken
	}

	if token.IsRevoked() {
		return nil, tokensDomain.ErrInvalidToken
	}

	if token.ExpiresAt().Before(now) {
		return nil, tokensDomain.ErrInvalidToken
	}

	if err := s.refreshRepo.Revoke(ctx, hash); err != nil {
		return nil, err
	}

	return s.generateTokens(ctx, token.UserID())
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash := s.tokenHasher.Hash(refreshToken)

	token, err := s.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return err
	}
	if token == nil {
		return nil
	}

	return s.refreshRepo.Revoke(ctx, hash)
}

func (s *authService) generateTokens(ctx context.Context, userID uuid.UUID) (*dto.TokenPair, error) {
	access, err := s.tokenGen.GenerateAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.tokenGen.GenerateRefresh(userID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	hash := s.tokenHasher.Hash(refresh)
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	token, err := tokensDomain.NewTokens(userID, hash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, token, sessionLimit); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	tokens, err := tokensDomain.NewTokens(userID, refresh, now.Add(s.accessTTL))
	if err != nil {
		return nil, err
	}

	return &dto.TokenPair{
		AccessToken:  access,
		RefreshToken: tokens.RefreshToken(),
	}, nil
}
