package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	tokenmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
	txmanagermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/txmanager"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/user"
	"github.com/stretchr/testify/suite"
)

type UserSyncService interface {
	Delete(ctx context.Context, userID uuid.UUID) error
	Ban(ctx context.Context, userID uuid.UUID) error
	Unban(ctx context.Context, userID uuid.UUID) error
}

type UserSyncServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service UserSyncService

	userRepo  userDomain.UserRepository
	tokenRepo tokenDomain.TokenRepository

	txManager port.TxManager

	passwordHasher port.PasswordHasher
}

func (s *UserSyncServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.userRepo = usermem.NewUserRepository()
	s.tokenRepo = tokenmem.NewTokenRepository()

	s.passwordHasher = &security.FakeHasher{}
	s.txManager = txmanagermem.NewFakeTxManager()

	s.service = user.NewUserSyncService(s.txManager, s.userRepo, s.tokenRepo)
}

func (s *UserSyncServiceTestSuite) TestDeleteSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.Delete(s.ctx, userID)
	s.Require().NoError(err)

	updated, err := s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	expectedStatus, err := userDomain.ParseStatus("inactive")
	s.Require().NoError(err)

	s.Require().Equal(expectedStatus, updated.Status())
}

func (s *UserSyncServiceTestSuite) TestDeleteInvalidUserID() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.Delete(s.ctx, uuid.Nil)
	s.Require().Error(err)
}

func (s *UserSyncServiceTestSuite) TestDeleteUserNotFoundIgnoreErr() {
	userID := uuid.New()

	err := s.service.Delete(s.ctx, userID)
	s.Require().NoError(err)
}

// Ban
func (s *UserSyncServiceTestSuite) TestBanSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.Ban(s.ctx, userID)
	s.Require().NoError(err)

	updated, err := s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	expectedStatus, err := userDomain.ParseStatus("banned")
	s.Require().NoError(err)

	s.Require().Equal(expectedStatus, updated.Status())
}

func (s *UserSyncServiceTestSuite) TestBanInvalidUserID() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.Ban(s.ctx, uuid.Nil)
	s.Require().Error(err)
}

func (s *UserSyncServiceTestSuite) TestBanUserNotFoundIgnoreErr() {
	userID := uuid.New()

	err := s.service.Ban(s.ctx, userID)
	s.Require().NoError(err)
}

// Unban
func (s *UserSyncServiceTestSuite) TestUnbanSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.Ban(s.ctx, userID)
	s.Require().NoError(err)

	updated, err := s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	expectedStatus, err := userDomain.ParseStatus("banned")
	s.Require().NoError(err)

	s.Require().Equal(expectedStatus, updated.Status())

	err = s.service.Unban(s.ctx, userID)
	s.Require().NoError(err)

	updated, err = s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	expectedStatus, err = userDomain.ParseStatus("active")
	s.Require().NoError(err)

	s.Require().Equal(expectedStatus, updated.Status())
}

func TestUserSyncServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserSyncServiceTestSuite))
}
