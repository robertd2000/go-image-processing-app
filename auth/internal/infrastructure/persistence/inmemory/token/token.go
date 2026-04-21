// Package tokenmem
package tokenmem

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type tokenInMemoryRepository struct {
	mu   *sync.RWMutex
	data map[string]*tokenDomain.Tokens
}

func NewTokenRepository() tokenDomain.TokenRepository {
	return &tokenInMemoryRepository{
		data: make(map[string]*tokenDomain.Tokens),
		mu:   &sync.RWMutex{},
	}
}

func (r *tokenInMemoryRepository) Create(ctx context.Context, token *tokenDomain.Tokens, limit int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existedToken, _ := r.getByToken(ctx, token.RefreshToken())
	if existedToken != nil {
		return tokenDomain.ErrTokenAlreadyExists
	}

	r.data[token.RefreshToken()] = token

	return nil
}

func (t *tokenInMemoryRepository) IsValid(ctx context.Context, userID uuid.UUID, token string) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tokens, err := t.getByToken(ctx, token)
	if err != nil {
		if errors.Is(err, tokenDomain.ErrTokenNotFound) {
			return false, nil
		}

		return false, err
	}

	if tokens.UserID() != userID {
		return false, nil
	}

	if tokens.IsRevoked() {
		return false, nil
	}

	if tokens.IsExpired(time.Now()) {
		return false, nil
	}

	return true, nil
}

func (t *tokenInMemoryRepository) Update(ctx context.Context, userID uuid.UUID, oldToken, newToken string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tokens, err := t.getByToken(ctx, oldToken)
	if err != nil {
		return err
	}

	if tokens.UserID() != userID {
		return tokenDomain.ErrTokenNotFound
	}

	if _, exists := t.data[newToken]; exists {
		return tokenDomain.ErrTokenAlreadyExists
	}

	delete(t.data, oldToken)

	newEntity := tokenDomain.RestoreTokens(
		tokens.ID(),
		tokens.UserID(),
		newToken,
		tokens.ExpiresAt(),
		tokens.CreatedAt(),
		tokens.RevokedAt(),
		tokens.FamilyID(),
		tokens.ParentID(),
	)

	t.data[newToken] = newEntity

	return nil
}

func (t *tokenInMemoryRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	found := false
	for _, tokens := range t.data {
		if tokens == nil {
			continue
		}
		if tokens.ID() == tokenID {
			now := time.Now().UTC()
			tokens.Revoke(now)
			found = true
			break
		}
	}

	if !found {
		return tokenDomain.ErrTokenNotFound
	}

	return nil
}

func (t *tokenInMemoryRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now().UTC()

	for _, tokens := range t.data {
		if tokens == nil {
			continue
		}
		if tokens.FamilyID() == familyID {
			tokens.Revoke(now)
		}
	}

	return nil
}

func (t *tokenInMemoryRepository) Rotate(
	ctx context.Context,
	oldToken *tokenDomain.Tokens,
	newToken *tokenDomain.Tokens,
) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	existedToken, err := t.getByToken(ctx, oldToken.RefreshToken())
	if err != nil {
		return false, err
	}

	if existedToken.ID() != oldToken.ID() {
		return false, tokenDomain.ErrTokenNotFound
	}

	if existedToken.IsRevoked() {
		return true, nil
	}

	existedToken.Revoke(time.Now().UTC())

	if _, exists := t.data[newToken.RefreshToken()]; exists {
		return false, tokenDomain.ErrTokenAlreadyExists
	}

	t.data[newToken.RefreshToken()] = newToken

	return false, nil
}

func (t *tokenInMemoryRepository) GetByHash(ctx context.Context, hash string) (*tokenDomain.Tokens, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.getByToken(ctx, hash)
}

func (t *tokenInMemoryRepository) getByToken(_ context.Context, token string) (*tokenDomain.Tokens, error) {
	tokens, exists := t.data[token]
	if !exists {
		return nil, tokenDomain.ErrTokenNotFound
	}

	return tokens, nil
}

func (r *tokenInMemoryRepository) DeleteByUserID(ctx context.Context, tx port.Tx, userID uuid.UUID) error {
	return nil
}
