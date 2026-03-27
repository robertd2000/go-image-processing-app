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
	id           uuid.UUID
	userID       uuid.UUID
	refreshToken string
	expiresAt    time.Time
	createdAt    time.Time
	revokedAt    *time.Time
	parentID     uuid.UUID // Optional: ID of the parent token for family revocation
	familyID     uuid.UUID
}

func NewTokens(
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
	familyID uuid.UUID,
	parentID uuid.UUID,
) (*Tokens, error) {

	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if err := validateToken(refreshToken); err != nil {
		return nil, err
	}

	id := uuid.New()

	if familyID == uuid.Nil {
		familyID = id
	}

	return &Tokens{
		id:           id,
		userID:       userID,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		createdAt:    time.Now(),
		familyID:     familyID,
		parentID:     parentID,
	}, nil
}

func RestoreTokens(
	id uuid.UUID,
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
	createdAt time.Time,
	revokedAt *time.Time,
	familyID uuid.UUID,
	parentID uuid.UUID,
) *Tokens {
	return &Tokens{
		id:           id,
		userID:       userID,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		createdAt:    createdAt,
		revokedAt:    revokedAt,
		familyID:     familyID,
		parentID:     parentID,
	}
}

func (t *Tokens) ID() uuid.UUID {
	return t.id
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

func (t *Tokens) RevokedAt() *time.Time {
	return t.revokedAt
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

func (t *Tokens) FamilyID() uuid.UUID {
	return t.familyID
}

func (t *Tokens) SetFamilyID(familyID uuid.UUID) {
	t.familyID = familyID
}

func (t *Tokens) ParentID() uuid.UUID {
	return t.parentID
}

func (t *Tokens) SetParentID(parentID uuid.UUID) {
	t.parentID = parentID
}

func validateToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	return nil
}
