package jwt

import (
	"errors"

	"github.com/google/uuid"
)

type InMemoryTokenGenerator struct {
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

func (g *InMemoryTokenGenerator) Generate(userID uuid.UUID, email string) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "generic_" + userID.String()
	g.genericTokens[token] = userID
	return token, nil
}

func (g *InMemoryTokenGenerator) Validate(token string) (uuid.UUID, error) {
	if g.ValidateErr != nil {
		return uuid.Nil, g.ValidateErr
	}

	id, ok := g.genericTokens[token]
	if !ok {
		return uuid.Nil, errors.New("invalid token")
	}

	return id, nil
}

func (g *InMemoryTokenGenerator) GenerateAccess(userID uuid.UUID) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "access_" + userID.String()
	g.accessTokens[token] = userID
	return token, nil
}

func (g *InMemoryTokenGenerator) GenerateRefresh(userID uuid.UUID) (string, error) {
	if g.GenerateErr != nil {
		return "", g.GenerateErr
	}

	token := "refresh_" + userID.String()
	g.refreshTokens[token] = userID
	return token, nil
}

func (g *InMemoryTokenGenerator) ValidateAccess(token string) (uuid.UUID, error) {
	if g.ValidateErr != nil {
		return uuid.Nil, g.ValidateErr
	}

	id, ok := g.accessTokens[token]
	if !ok {
		return uuid.Nil, errors.New("invalid access token")
	}

	return id, nil
}

func (g *InMemoryTokenGenerator) ValidateRefresh(token string) (uuid.UUID, error) {
	if g.ValidateErr != nil {
		return uuid.Nil, g.ValidateErr
	}

	id, ok := g.refreshTokens[token]
	if !ok {
		return uuid.Nil, errors.New("invalid refresh token")
	}

	return id, nil
}
