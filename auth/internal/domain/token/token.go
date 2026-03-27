package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidRefresh       = errors.New("invalid refresh")
	ErrTokenAlreadyExists   = errors.New("token already exists")
	ErrTokenNotFound        = errors.New("token not found")
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrExpiredToken         = errors.New("token has expired")
	ErrSessionLimitExceeded = errors.New("session limit exceeded")
)

type Tokens struct {
	userID       uuid.UUID
	refreshToken string
	expiresAt    time.Time
	createdAt    time.Time
	revokedAt    *time.Time
}

func NewTokens(
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
) (*Tokens, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if err := validateToken(refreshToken); err != nil {
		return nil, err
	}

	return &Tokens{
		userID:       userID,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		createdAt:    time.Now(),
	}, nil
}

func RestoreTokens(
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
	createdAt time.Time,
	revokedAt *time.Time,
) *Tokens {
	return &Tokens{
		userID:       userID,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		createdAt:    createdAt,
		revokedAt:    revokedAt,
	}
}

func (t *Tokens) UserID() uuid.UUID {
	return t.userID
}

func (t *Tokens) RefreshToken() string {
	return t.refreshToken
}

func (t *Tokens) IsRevoked() bool {
	return t.revokedAt != nil
}

func (t *Tokens) IsExpired(now time.Time) bool {
	return now.After(t.expiresAt)
}

func (t *Tokens) Revoke(now time.Time) {
	if t.revokedAt != nil {
		return
	}
	t.revokedAt = &now
}

func (t *Tokens) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *Tokens) CreatedAt() time.Time {
	return t.createdAt
}

func validateToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	return nil
}
