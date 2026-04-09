package jwt

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
)

type InMemoryTokenGenerator struct {
	mu sync.RWMutex

	accessTokens  map[string]uuid.UUID
	refreshTokens map[string]uuid.UUID
	genericTokens map[string]uuid.UUID

	GenerateErr error
	ValidateErr error
}

func NewInMemoryTokenGenerator() *InMemoryTokenGenerator {
	return &InMemoryTokenGenerator{
		accessTokens:  make(map[string]uuid.UUID),
		refreshTokens: make(map[string]uuid.UUID),
		genericTokens: make(map[string]uuid.UUID),
	}
}

func (g *InMemoryTokenGenerator) Generate(claims model.ClaimsInput) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "generic_" + uuid.NewString()

	g.mu.Lock()
	g.genericTokens[token] = claims.UserID
	g.mu.Unlock()

	return token, nil
}

func (g *InMemoryTokenGenerator) Validate(token string) (*auth.Claims, error) {
	if g.ValidateErr != nil {
		return nil, g.ValidateErr
	}

	g.mu.RLock()
	id, ok := g.genericTokens[token]
	g.mu.RUnlock()

	if !ok {
		return nil, errors.New("invalid token")
	}

	return &auth.Claims{UserID: id}, nil
}

func (g *InMemoryTokenGenerator) GenerateAccess(input model.ClaimsInput) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "access_" + uuid.NewString()

	g.mu.Lock()
	g.accessTokens[token] = input.UserID
	g.mu.Unlock()

	return token, nil
}

func (g *InMemoryTokenGenerator) GenerateRefresh(userID uuid.UUID) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "refresh_" + uuid.NewString()

	g.mu.Lock()
	g.refreshTokens[token] = userID
	g.mu.Unlock()

	return token, nil
}

func (g *InMemoryTokenGenerator) ValidateAccess(token string) (*auth.Claims, error) {
	if g.ValidateErr != nil {
		return nil, g.ValidateErr
	}

	g.mu.RLock()
	id, ok := g.accessTokens[token]
	g.mu.RUnlock()

	if !ok {
		return nil, errors.New("invalid access token")
	}

	return &auth.Claims{UserID: id}, nil
}

func (g *InMemoryTokenGenerator) ValidateRefresh(token string) (uuid.UUID, error) {
	if g.ValidateErr != nil {
		return uuid.Nil, g.ValidateErr
	}

	g.mu.RLock()
	id, ok := g.refreshTokens[token]
	g.mu.RUnlock()

	if !ok {
		return uuid.Nil, errors.New("invalid refresh token")
	}

	return id, nil
}
