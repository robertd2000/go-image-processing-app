package auth_test

import (
	"context"
	"testing"

	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	tokenmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	AuthService interface {
		Register(ctx context.Context, username, fistname, lastname, email, password string) error
		Login(ctx context.Context, email string, password string) (tokensDomain.Tokens, error)
		Refresh(ctx context.Context, refreshToken string) (tokensDomain.Tokens, error)
		Logout(ctx context.Context, refreshToken string) error
	}

	AuthTestSuite struct {
		suite.Suite

		ctx context.Context

		service AuthService

		hasher    port.PasswordHasher
		userRepo  user.UserRepository
		tokenRepo tokensDomain.TokenRepository
		tokenGen  port.TokenGenerator
	}
)

func (s *AuthTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.hasher = &security.FakeHasher{}
	s.tokenGen = jwt.NewInMemoryTokenGenerator()
	s.userRepo = usermem.NewUserRepository()
	s.tokenRepo = tokenmem.NewTokenRepository()

	s.service = auth.NewAuthService(s.userRepo, s.tokenRepo, s.hasher, s.tokenGen)
}

func (s *AuthTestSuite) TestAuthService_Register_Success() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	err := s.service.Register(
		s.ctx,
		username,
		firstname,
		lastname,
		email,
		password,
	)

	s.Require().NoError(err)

	userEntity, err := s.userRepo.GetByEmail(s.ctx, email)
	s.Require().NoError(err)
	s.Require().NotNil(userEntity)
}

func (s *AuthTestSuite) TestAuthService_Register_UserAlreadyExists() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	err := s.service.Register(s.ctx, username, firstname, lastname, email, password)
	assert.NoError(s.T(), err)

	err = s.service.Register(s.ctx, username, firstname, lastname, email, password)
	assert.ErrorIs(s.T(), err, user.ErrUserAlreadyExists)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidEmail() {
	ctx := context.Background()
	password := "!Secure123"
	email := "test_user1"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	err := s.service.Register(ctx, username, firstname, lastname, email, password)
	assert.ErrorIs(s.T(), err, user.ErrInvalidEmail)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidPassword() {
	err := s.service.Register(
		s.ctx,
		"test_user",
		"user",
		"1",
		"test@example.com",
		"",
	)

	assert.Error(s.T(), err)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
