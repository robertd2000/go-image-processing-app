// Package auth
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/validation"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

var sessionLimit = 5

type authService struct {
	userRepo    userDomain.UserRepository
	refreshRepo tokensDomain.TokenRepository
	outboxRepo  port.OutboxRepository

	tokenGen       port.TokenGenerator
	passwordHasher port.PasswordHasher
	tokenHasher    port.TokenHasher
	txManager      port.TxManager

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	userRepo userDomain.UserRepository,
	refreshRepo tokensDomain.TokenRepository,
	outboxRepo port.OutboxRepository,
	passwordHasher port.PasswordHasher,
	tokenHasher port.TokenHasher,
	tokenGen port.TokenGenerator,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	txManager port.TxManager,
) *authService {
	return &authService{
		userRepo:       userRepo,
		refreshRepo:    refreshRepo,
		outboxRepo:     outboxRepo,
		tokenGen:       tokenGen,
		passwordHasher: passwordHasher,
		tokenHasher:    tokenHasher,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		txManager:      txManager,
	}
}

func (s *authService) Register(ctx context.Context, in model.RegisterInput) error {
	if err := validation.ValidateEmail(in.Email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(in.Password); err != nil {
		return err
	}

	if err := validation.ValidateUsername(in.Username); err != nil {
		return err
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return err
	}
	if exists {
		return userDomain.ErrUserAlreadyExists
	}

	hashed, err := s.passwordHasher.Hash(in.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := userDomain.NewAuthUser(uuid.New(), in.Username, &in.Email, hashed)
	if err != nil {
		return err
	}

	return s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {

		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return err
		}
		eventID := uuid.New()

		event, err := events.NewEvent(
			"user.created",
			1,
			events.UserCreatedEvent{
				ID:        user.ID(),
				Username:  user.Username(),
				Email:     *user.Email(),
				CreatedAt: user.CreatedAt(),
			},
		)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}

		outboxEvent := port.OutboxEvent{
			ID:        eventID,
			Type:      "user.created",
			Topic:     "user.events.v1",
			Key:       user.ID().String(),
			Payload:   payload,
			CreatedAt: time.Now(),
		}

		if err := s.outboxRepo.Create(ctx, tx, outboxEvent); err != nil {
			return err
		}

		return nil
	})
}

func (s *authService) Login(ctx context.Context, in model.LoginInput) (*model.TokenPair, error) {
	if err := validation.ValidateEmail(in.Email); err != nil {
		return nil, err
	}

	if err := validation.ValidatePassword(in.Password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, userDomain.ErrUserNotFound) {
			return nil, userDomain.ErrWrongCredentials
		}
		return nil, err
	}

	if !user.Status().IsActive() {
		return nil, userDomain.ErrUserDisabled
	}

	if !s.passwordHasher.Compare(in.Password, user.PasswordHash()) {
		return nil, userDomain.ErrWrongCredentials
	}

	return s.generateTokenPair(ctx, user.ID())
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*model.TokenPair, error) {
	if refreshToken == "" {
		return nil, tokensDomain.ErrInvalidToken
	}
	now := time.Now()
	hash := s.tokenHasher.Hash(refreshToken)

	token, err := s.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, tokensDomain.ErrInvalidToken
	}

	if token.ExpiresAt().Before(now) {
		return nil, tokensDomain.ErrExpiredToken
	}

	user, err := s.userRepo.GetByID(ctx, token.UserID())
	if err != nil {
		return nil, err
	}

	var roles []string
	for _, r := range user.Roles() {
		roles = append(roles, r.Name().String())
	}

	access, err := s.tokenGen.GenerateAccess(model.ClaimsInput{
		UserID: user.ID(),
		Roles:  roles,
	})
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
		now.Add(s.refreshTTL),
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

	return &model.TokenPair{
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

func (s *authService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*model.TokenPair, error) {
	user, err := s.userRepo.GetByID(ctx, userID)

	var roles []string
	for _, r := range user.Roles() {
		roles = append(roles, r.Name().String())
	}

	access, err := s.tokenGen.GenerateAccess(model.ClaimsInput{
		UserID: user.ID(),
		Roles:  roles,
	})
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

	return &model.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
