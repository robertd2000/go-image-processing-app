// Package auth
package auth

import (
	"context"
	"errors"
	"fmt"

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

func (s *authService) Register(ctx context.Context, username, fistname, lastname, email, password string) error {
	if err := validation.ValidateEmail(email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return err
	}

	if err := validation.ValidateUsername(username); err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("error by finding user by email")
	}
	if exists {
		return userDomain.ErrUserAlreadyExists
	}

	hashed, err := s.hasher.Hash(password)
	if err != nil {
		if errors.Is(err, userDomain.ErrInvalidPasswordHash) {
			return err
		}
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := userDomain.CreateUser(username, fistname, lastname, &email, hashed)
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

func (s *authService) Login(ctx context.Context, email string, password string) (tokensDomain.Tokens, error) {
	return tokensDomain.Tokens{}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (tokensDomain.Tokens, error) {
	return tokensDomain.Tokens{}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}
