package auth_test

import (
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	userInmemory "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
)

type (
	AuthService interface{}

	AuthTestSuite struct {
		service   AuthService
		hasher    port.PasswordHasher
		userRepo  user.UserRepository
		tokenRepo token.TokenRepository
		tokenGen  port.TokenGenerator
	}
)

func (s *AuthTestSuite) SetupTest() {
	s.hasher = &security.FakeHasher{}
	s.tokenGen = jwt.NewInMemoryTokenGenerator()
	s.userRepo = userInmemory.NewUserInMemoryRepository()

	s.service = auth.NewAuthService(s.userRepo, s.tokenRepo, s.hasher, s.tokenGen)
}
