package auth_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	eventpub "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/events"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	tokenmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	AuthService interface {
		Register(ctx context.Context, in model.RegisterInput) error
		Login(ctx context.Context, in model.LoginInput) (*model.TokenPair, error)
		Refresh(ctx context.Context, refreshToken string) (*model.TokenPair, error)
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
		eventPublisher port.EventPublisher
	}
)

func (s *AuthTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.tokenGen = jwt.NewInMemoryTokenGenerator()
	s.userRepo = usermem.NewUserRepository()
	s.tokenRepo = tokenmem.NewTokenRepository()
	s.passwordHasher = &security.FakeHasher{}
	s.tokenHasher = &security.FakeTokenHasher{}
	s.eventPublisher = &eventpub.FakeEventPublisher{}

	s.service = auth.NewAuthService(
		s.userRepo,
		s.tokenRepo,
		s.passwordHasher,
		s.tokenHasher,
		s.tokenGen,
		s.eventPublisher,
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

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}

	err := s.service.Register(
		s.ctx,
		registerInput,
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

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}

	err := s.service.Register(s.ctx, registerInput)
	assert.NoError(s.T(), err)

	registerInput.Username = "test_user2"
	err = s.service.Register(s.ctx, registerInput)
	assert.ErrorIs(s.T(), err, userDomain.ErrUserAlreadyExists)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidEmail() {
	ctx := context.Background()
	password := "!Secure123"
	email := "test_user1"
	username := "test_user"
	firstname := "user"
	lastname := "1"

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}

	err := s.service.Register(ctx, registerInput)
	assert.ErrorIs(s.T(), err, userDomain.ErrInvalidEmail)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidUsername() {
	ctx := context.Background()
	password := "!Secure123"
	email := "test_user1@example.com"
	username := ""
	firstname := "user"
	lastname := "1"

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}
	err := s.service.Register(ctx, registerInput)
	assert.ErrorIs(s.T(), err, userDomain.ErrInvalidUsername)
}

func (s *AuthTestSuite) TestAuthService_Register_InvalidPassword() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := ""
	firstname := "user"
	lastname := "1"

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}
	err := s.service.Register(
		s.ctx,
		registerInput,
	)

	assert.Error(s.T(), err)
}

func (s *AuthTestSuite) TestAuthService_LoginSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(uuid.New(), username, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)
	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	s.NotEmpty(tokens.AccessToken)
	s.NotEmpty(tokens.RefreshToken)
}

func (s *AuthTestSuite) TestAuthService_LoginUserNotExists() {
	password := "!Secure123"
	email := "test_user1@example.com"

	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrWrongCredentials)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginInvalidEmail() {
	password := "!Secure123"
	email := "test_user1"

	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrInvalidEmail)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginInvalidPasswordFormat() {
	password := "!Secure"
	email := "test_user1@example"

	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrInvalidPassword)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginWrongPassword() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(uuid.New(), username, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)
	loginInput := model.LoginInput{
		Email:    email,
		Password: "!!!!!!SecureDifferent22",
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
	s.Require().Error(err)
	s.Require().ErrorIs(err, userDomain.ErrWrongCredentials)
	s.Require().Nil(tokens)
}

func (s *AuthTestSuite) TestAuthService_LoginDisabledUser() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"

	hashed, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(uuid.New(), username, &email, hashed)
	s.Require().NoError(err)

	err = s.userRepo.Create(s.ctx, user)
	s.Require().NoError(err)

	err = s.userRepo.Disable(s.ctx, user.ID())
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(s.ctx, loginInput)
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

	registerInput := model.RegisterInput{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstname,
		LastName:  lastname,
	}

	err := s.service.Register(ctx, registerInput)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    email,
		Password: password,
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)

	newTokens, err := s.service.Refresh(ctx, tokens.RefreshToken)

	s.Require().NoError(err)
	s.Require().NotNil(newTokens)

	s.NotEqual(tokens.AccessToken, newTokens.AccessToken)
	s.NotEqual(tokens.RefreshToken, newTokens.RefreshToken)
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

	registerInput := model.RegisterInput{
		Username:  "test_user",
		Email:     "john2@example.com",
		Password:  "!Secure123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := s.service.Register(
		ctx,
		registerInput,
	)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    "john2@example.com",
		Password: "!Secure123",
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)

	// Extract the token entity from the repository to get its ID
	hashedRefreshToken := s.tokenHasher.Hash(tokens.RefreshToken)
	refreshTokenEntity, err := s.tokenRepo.GetByHash(ctx, hashedRefreshToken)
	s.Require().NoError(err)

	err = s.tokenRepo.Revoke(ctx, refreshTokenEntity.ID())
	_, err = s.service.Refresh(ctx, tokens.RefreshToken)

	s.Require().Error(err)
	s.Require().ErrorIs(err, tokensDomain.ErrInvalidToken)
	_, err = s.service.Refresh(ctx, tokens.RefreshToken)
	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Logout_Success() {
	ctx := s.ctx

	registerInput := model.RegisterInput{
		Username:  "user_logout",
		Email:     "logout@example.com",
		Password:  "!Secure123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := s.service.Register(ctx, registerInput)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    "logout@example.com",
		Password: "!Secure123",
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken)
	s.Require().NoError(err)
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

	registerInput := model.RegisterInput{
		Username:  "user_logout2",
		Email:     "logout2@example.com",
		Password:  "!Secure123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := s.service.Register(ctx, registerInput)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    "logout2@example.com",
		Password: "!Secure123",
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken)
	s.Require().NoError(err)

	err = s.service.Logout(ctx, tokens.RefreshToken)
	s.Require().NoError(err)
}

func (s *AuthTestSuite) TestAuthService_Refresh_ReuseAttack_ShouldRevokeFamily() {
	ctx := s.ctx

	registerInput := model.RegisterInput{
		Username:  "user_logout2",
		Email:     "logout2@example.com",
		Password:  "!Secure123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := s.service.Register(ctx, registerInput)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    "logout2@example.com",
		Password: "!Secure123",
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)

	newTokens, err := s.service.Refresh(ctx, tokens.RefreshToken)
	s.Require().NoError(err)
	s.Require().NotNil(newTokens)
	s.Require().NotEmpty(newTokens.RefreshToken)
	fmt.Println("DEBUG newTokens:", newTokens)
	_, err = s.service.Refresh(ctx, tokens.RefreshToken)
	s.Require().Error(err)

	_, err = s.service.Refresh(ctx, newTokens.RefreshToken)
	s.Require().Error(err)
}

func (s *AuthTestSuite) TestAuthService_Refresh_ShouldPreserveFamily() {
	ctx := s.ctx

	registerInput := model.RegisterInput{
		Username:  "user_logout2",
		Email:     "logout2@example.com",
		Password:  "!Secure123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := s.service.Register(
		ctx,
		registerInput,
	)
	s.Require().NoError(err)

	loginInput := model.LoginInput{
		Email:    "logout2@example.com",
		Password: "!Secure123",
	}
	tokens, err := s.service.Login(ctx, loginInput)
	s.Require().NoError(err)

	hash := s.tokenHasher.Hash(tokens.RefreshToken)
	token1, _ := s.tokenRepo.GetByHash(ctx, hash)

	newTokens, _ := s.service.Refresh(ctx, tokens.RefreshToken)

	hash2 := s.tokenHasher.Hash(newTokens.RefreshToken)
	token2, _ := s.tokenRepo.GetByHash(ctx, hash2)

	s.Require().Equal(token1.FamilyID(), token2.FamilyID())
	s.Require().Equal(token1.ID(), token2.ParentID())
}

func (s *AuthTestSuite) TestAuthService_Refresh_ExpiredToken() {
	ctx := s.ctx

	s.service = auth.NewAuthService(
		s.userRepo,
		s.tokenRepo,
		s.passwordHasher,
		s.tokenHasher,
		s.tokenGen,
		s.eventPublisher,
		1*time.Second,
		1*time.Second,
	)

	registerInput := model.RegisterInput{
		Username:  "exp",
		Email:     "exp@test.com",
		Password:  "!Secure123",
		FirstName: "f",
		LastName:  "l",
	}

	_ = s.service.Register(ctx, registerInput)
	loginInput := model.LoginInput{
		Email:    "exp@test.com",
		Password: "!Secure123",
	}
	tokens, _ := s.service.Login(ctx, loginInput)

	time.Sleep(2 * time.Second)

	_, err := s.service.Refresh(ctx, tokens.RefreshToken)

	s.Require().ErrorIs(err, tokensDomain.ErrExpiredToken)
}

func (s *AuthTestSuite) TestAuthService_Refresh_RaceCondition() {
	ctx := s.ctx

	registerInput := model.RegisterInput{
		Username:  "race",
		Email:     "race@test.com",
		Password:  "!Secure123",
		FirstName: "f",
		LastName:  "l",
	}

	_ = s.service.Register(ctx, registerInput)
	loginInput := model.LoginInput{
		Email:    "race@test.com",
		Password: "!Secure123",
	}
	tokens, _ := s.service.Login(ctx, loginInput)

	var wg sync.WaitGroup
	results := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := s.service.Refresh(ctx, tokens.RefreshToken)
		results <- err
	}()

	go func() {
		defer wg.Done()
		_, err := s.service.Refresh(ctx, tokens.RefreshToken)
		results <- err
	}()

	wg.Wait()
	close(results)

	success := 0
	fail := 0

	for err := range results {
		if err == nil {
			success++
		} else {
			fail++
		}
	}

	s.Require().Equal(1, success)
	s.Require().Equal(1, fail)
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
