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

func (r *tokenInMemoryRepository) Create(ctx context.Context, token *tokenDomain.Tokens, limit int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existedToken, _ := r.getByToken(ctx, token.RefreshToken())
	if existedToken != nil {
		return tokenDomain.ErrTokenAlreadyExists
	}

	tokens, err := tokenDomain.NewTokens(token.UserID(), token.RefreshToken()+"_refresh", token.ExpiresAt(), token.FamilyID(), token.ParentID())
	if err != nil {
		return err
	}

	r.data[token.RefreshToken()] = tokens

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

	for _, tokens := range t.data {
		if tokens.ID() == tokenID {
			now := time.Now().UTC()
			tokens.Revoke(now)
			break
		}
	}

	return nil
}

func (t *tokenInMemoryRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tokens := range t.data {
		if tokens.FamilyID() != uuid.Nil && tokens.FamilyID() == familyID {
			now := time.Now().UTC()
			tokens.Revoke(now)
		}
	}

	return nil
}

func (t *tokenInMemoryRepository) Rotate(ctx context.Context, oldToken *tokenDomain.Tokens, newToken *tokenDomain.Tokens) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tokens := range t.data {
		if tokens.ID() == oldToken.ID() {
			delete(t.data, tokens.RefreshToken())
			tokens.Revoke(time.Now().UTC())
			t.data[tokens.RefreshToken()] = tokens
			t.data[newToken.RefreshToken()] = newToken
			return nil
		}
	}

	return tokenDomain.ErrTokenNotFound
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
