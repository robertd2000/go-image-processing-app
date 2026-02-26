// Package auth
package auth

import (
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
)

type authService struct {
	userRepo    user.UserRepository
	refreshRepo token.TokenRepository
	hasher      port.PasswordHasher
	tokenGen    port.TokenGenerator
}

func NewAuthService(userRepo user.UserRepository, refreshRepo token.TokenRepository, hasher port.PasswordHasher, tokenGen port.TokenGenerator) *authService {
	return &authService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		hasher:      hasher,
		tokenGen:    tokenGen,
	}
}
