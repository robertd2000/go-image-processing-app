package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
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
		Login(ctx context.Context, email string, password string) (*tokensDomain.Tokens, error)
		Refresh(ctx context.Context, refreshToken string) (*tokensDomain.Tokens, error)
		Logout(ctx context.Context, refreshToken string) error
	}

	AuthTestSuite struct {
		suite.Suite

		ctx context.Context

		service AuthService

		userRepo       userDomain.UserRepository
		tokenRepo      tokensDomain.TokenRepository
		tokenGen       port.TokenGenerator
		passwordHasher port.PasswordHasher
		tokenHasher    port.TokenHasher
	}
)

func (s *AuthTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.tokenGen = jwt.NewInMemoryTokenGenerator()
	s.userRepo = usermem.NewUserRepository()
	s.tokenRepo = tokenmem.NewTokenRepository()
	s.passwordHasher = &security.FakeHasher{}
	s.tokenHasher = &security.FakeTokenHasher{}

	s.service = auth.NewAuthService(
		s.userRepo,
		s.tokenRepo,
		s.passwordHasher,
		s.tokenHasher,
		s.tokenGen,
		10*time.Minute,
		60*time.Minute,
	)
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
	assert.ErrorIs(s.T(), err, userDomain.ErrUserAlreadyExists)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidEmail() {
	ctx := context.Background()
	password := "!Secure123"
	email := "test_user1"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	err := s.service.Register(ctx, username, firstname, lastname, email, password)
	assert.ErrorIs(s.T(), err, userDomain.ErrInvalidEmail)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidUsername() {
	ctx := context.Background()
	password := "!Secure123"
	email := "test_user1@example.com"
	username := ""
	firstname := "user"
	lastname := "1"

	err := s.service.Register(ctx, username, firstname, lastname, email, password)
	assert.ErrorIs(s.T(), err, userDomain.ErrInvalidUsername)
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

func (s *AuthTestSuite) TestAuthService_LoginSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.CreateUser(username, firstname, lastname, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)
	tokens, err := s.service.Login(s.ctx, email, password)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.NotEmpty(tokens.AccessToken())
	s.NotEmpty(tokens.RefreshToken())
}

func (s *AuthTestSuite) TestAuthService_LoginUserNotExists() {
	password := "!Secure123"
	email := "test_user1@example.com"

	tokens, err := s.service.Login(s.ctx, email, password)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrWrongCreadentials)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginInvalidEmail() {
	password := "!Secure123"
	email := "test_user1"

	tokens, err := s.service.Login(s.ctx, email, password)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrInvalidEmail)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginInvalidPasswordFormat() {
	password := "!Secure"
	email := "test_user1@example"

	tokens, err := s.service.Login(s.ctx, email, password)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrInvalidPassword)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginWrongPassword() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.CreateUser(username, firstname, lastname, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)
	tokens, err := s.service.Login(s.ctx, email, "!!!!!!SecureDifferent22")
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrWrongCreadentials)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginDisabledUser() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.CreateUser(username, firstname, lastname, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)

	err = s.userRepo.Disable(s.ctx, user.ID())
	s.Require().NoError(err)

	tokens, err := s.service.Login(s.ctx, email, password)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrUserDisabled)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_Refresh_Success() {
	ctx := s.ctx

	password := "!Secure123"
	email := "refresh_success@example.com"
	username := "user1"
	firstname := "user"
	lastname := "one"

	err := s.service.Register(ctx, username, firstname, lastname, email, password)
	s.Require().NoError(err)

	tokens, err := s.service.Login(ctx, email, password)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	newTokens, err := s.service.Refresh(ctx, tokens.RefreshToken())

	s.Require().NoError(err)
	s.Require().NotNil(newTokens)

	s.NotEqual(tokens.AccessToken(), newTokens.AccessToken())
	s.NotEqual(tokens.RefreshToken(), newTokens.RefreshToken())
}

func (s *AuthTestSuite) TestAuthService_Refresh_InvalidToken() {
	ctx := s.ctx

	_, err := s.service.Refresh(ctx, "invalid_token")

	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Refresh_TokenNotInRepo() {
	ctx := s.ctx

	userID := uuid.New()

	refresh, err := s.tokenGen.GenerateRefresh(userID)
	s.Require().NoError(err)

	_, err = s.service.Refresh(ctx, refresh)

	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Refresh_TokenRevoked() {
	ctx := s.ctx

	err := s.service.Register(
		ctx,
		"test_user",
		"John",
		"Doe",
		"john2@example.com",
		"!Secure123",
	)
	s.Require().NoError(err)

	tokens, err := s.service.Login(ctx, "john2@example.com", "!Secure123")
	s.Require().NoError(err)

	userID, err := s.tokenGen.ValidateRefresh(tokens.RefreshToken())
	s.Require().NoError(err)

	hash := s.tokenHasher.Hash(tokens.RefreshToken())

	err = s.tokenRepo.Revoke(ctx, userID, hash)
	s.Require().NoError(err)

	_, err = s.service.Refresh(ctx, tokens.RefreshToken())
	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Logout_Success() {
	ctx := s.ctx

	err := s.service.Register(
		ctx,
		"user_logout",
		"John",
		"Doe",
		"logout@example.com",
		"!Secure123",
	)
	s.Require().NoError(err)

	tokens, err := s.service.Login(ctx, "logout@example.com", "!Secure123")
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken())
	s.Require().NoError(err)

	userID, err := s.tokenGen.ValidateRefresh(tokens.RefreshToken())
	s.Require().NoError(err)

	ok, err := s.tokenRepo.IsValid(ctx, userID, tokens.RefreshToken())
	s.Require().NoError(err)
	s.False(ok)
}

func (s *AuthTestSuite) TestAuthService_Logout_InvalidToken() {
	ctx := s.ctx

	err := s.service.Logout(ctx, "invalid_token")

	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Logout_TokenNotInRepo() {
	ctx := s.ctx

	userID := uuid.New()

	refresh, err := s.tokenGen.GenerateRefresh(userID)
	s.Require().NoError(err)

	err = s.service.Logout(ctx, refresh)

	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Logout_AlreadyRevoked() {
	ctx := s.ctx

	err := s.service.Register(
		ctx,
		"user_logout2",
		"John",
		"Doe",
		"logout2@example.com",
		"!Secure123",
	)
	s.Require().NoError(err)

	tokens, err := s.service.Login(ctx, "logout2@example.com", "!Secure123")
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken())
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken())
	s.Require().NoError(err)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
