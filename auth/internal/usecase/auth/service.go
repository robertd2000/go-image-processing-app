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
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, username, firstName, lastName, email, password string) error {
	if err := validation.ValidateEmail(email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return err
	}

	if err := validation.ValidateUsername(username); err != nil {
		return err
	}
	hashed, err := s.passwordHasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := userDomain.CreateUser(username, firstName, lastName, &email, hashed)
	if err != nil {
		return err
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
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
		return nil, userDomain.ErrWrongCredentials
	}

	if !user.Enabled() {
		return nil, userDomain.ErrUserDisabled
	}

	if !s.passwordHasher.Compare(password, user.PasswordHash()) {
		return nil, userDomain.ErrWrongCredentials
	}

	return s.generateTokenPair(ctx, user.ID())
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error) {
	if refreshToken == "" {
		return nil, tokensDomain.ErrInvalidToken
	}
	hash := s.tokenHasher.Hash(refreshToken)

	token, err := s.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, tokensDomain.ErrInvalidToken
	}

	if token.ExpiresAt().Before(time.Now()) {
		return nil, tokensDomain.ErrExpiredToken
	}

	access, err := s.tokenGen.GenerateAccess(token.UserID())
	if err != nil {
		return nil, err
	}

	refresh, err := s.tokenGen.GenerateRefresh(token.UserID())
	if err != nil {
		return nil, err
	}
	familyID := token.FamilyID()
	if familyID == uuid.Nil {
		familyID = token.ID()
	}
	newToken, err := tokensDomain.NewTokens(
		token.UserID(),
		s.tokenHasher.Hash(refresh),
		time.Now().Add(s.refreshTTL),
		familyID,
		token.ID(),
	)
	if err != nil {
		return nil, err
	}
	if newToken == nil {
		return nil, errors.New("newToken is nil")
	}
	reuse, err := s.refreshRepo.Rotate(ctx, token, newToken)
	if err != nil {
		return nil, err
	}

	if reuse {
		_ = s.refreshRepo.RevokeFamily(ctx, token.FamilyID())
		return nil, tokensDomain.ErrInvalidToken
	}

	return &dto.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
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

	return s.refreshRepo.Revoke(ctx, token.ID())
}

func (s *authService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*dto.TokenPair, error) {
	access, err := s.tokenGen.GenerateAccess(userID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	if access == "" {
		return nil, fmt.Errorf("generate access token: empty token returned")
	}

	refresh, err := s.tokenGen.GenerateRefresh(userID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Cache the hash of the refresh token for reuse
	refreshHash := s.tokenHasher.Hash(refresh)
	now := time.Now()
	expiresAt := now.Add(s.refreshTTL)

	familyID := uuid.New()
	var parentID uuid.UUID

	token, err := tokensDomain.NewTokens(userID, refreshHash, expiresAt, familyID, parentID)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, token, sessionLimit); err != nil {
		if errors.Is(err, tokensDomain.ErrSessionLimitExceeded) {
			return nil, fmt.Errorf("session limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &dto.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
