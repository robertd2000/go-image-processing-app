package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/user"
	"github.com/stretchr/testify/suite"
)

type UserSyncService interface {
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
}

type UserSyncServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service UserSyncService

	userRepo userDomain.UserRepository

	passwordHasher port.PasswordHasher
}

func (s *UserSyncServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.userRepo = usermem.NewUserRepository()
	s.passwordHasher = &security.FakeHasher{}

	s.service = user.NewUserSyncService(s.userRepo)
}

func (s *UserSyncServiceTestSuite) TestUpdateStatusSuccess() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.UpdateStatus(s.ctx, userID, "inactive")
	s.Require().NoError(err)

	updated, err := s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	s.Require().Equal("inactive", updated.Status())
}

func (s *UserSyncServiceTestSuite) TestUpdateStatusInvalidUserID() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.UpdateStatus(s.ctx, uuid.Nil, "inactive")
	s.Require().Error(err)
}

func (s *UserSyncServiceTestSuite) TestUpdateStatusUserNotFoundIgnoreErr() {
	userID := uuid.New()

	err := s.service.UpdateStatus(s.ctx, userID, "inactive")
	s.Require().NoError(err)
}

func (s *UserSyncServiceTestSuite) TestUpdateStatusIgnoreIfSameStatus() {
	password := "!Secure123"
	email := "test_user1@example.com"
	username := "test_user"
	userID := uuid.New()
	passwordHash, err := s.passwordHasher.Hash(password)
	s.Require().NoError(err)

	user, err := userDomain.NewAuthUser(userID, username, &email, passwordHash)
	s.Require().NoError(err)
	s.Require().NoError(s.userRepo.Create(s.ctx, nil, user))

	err = s.service.UpdateStatus(s.ctx, userID, "active")
	s.Require().NoError(err)

	updated, err := s.userRepo.GetByID(s.ctx, userID)
	s.Require().NoError(err)

	s.Require().Equal("active", updated.Status())
}

func TestUserSyncServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserSyncServiceTestSuite))
}
