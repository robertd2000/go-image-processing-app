package tokenmem

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
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

func (t *tokenInMemoryRepository) Save(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	existedToken, _ := t.getByToken(ctx, token)
	if existedToken != nil {
		return tokenDomain.ErrTokenAlreadyExists
	}

	tokens, err := tokenDomain.NewTokens(userID, token, token+"_refresh", expiresAt)
	if err != nil {
		return err
	}

	t.data[token] = tokens

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
		userID,
		newToken,
		tokens.RefreshToken(),
		tokens.ExpiresAt(),
		tokens.CreatedAt(),
		nil,
	)

	t.data[newToken] = newEntity

	return nil
}

func (t *tokenInMemoryRepository) Revoke(ctx context.Context, userID uuid.UUID, token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tokens, err := t.getByToken(ctx, token)
	if err != nil {
		return err
	}

	if tokens.UserID() != userID {
		return tokenDomain.ErrTokenNotFound
	}

	now := time.Now().UTC()
	tokens.Revoke(now)

	return nil
}

func (t *tokenInMemoryRepository) RevokeByToken(ctx context.Context, token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tokens, err := t.getByToken(ctx, token)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	tokens.Revoke(now)

	return nil
}

func (t *tokenInMemoryRepository) GetByToken(ctx context.Context, token string) (*tokenDomain.Tokens, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.getByToken(ctx, token)
}

func (t *tokenInMemoryRepository) getByToken(ctx context.Context, token string) (*tokenDomain.Tokens, error) {
	tokens, exists := t.data[token]
	if !exists {
		return nil, tokenDomain.ErrTokenNotFound
	}

	return tokens, nil
}
