package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	usermem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserService interface {
	// Define methods for the UserService interface here
	Create(ctx context.Context, userInput model.CreateUserInput) error
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error)
}

type UserServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service UserService

	userRepo userDomain.UserRepository
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.userRepo = usermem.NewUserRepository()
	s.service = user.NewUserService(s.userRepo)
}

func (s *UserServiceTestSuite) TestCreateUser() {
	ctx := s.ctx
	userID := uuid.New()
	userName, err := userDomain.NewUsername("Test user")
	assert.NoError(s.T(), err)
	email, err := userDomain.NewEmail("test@example.com")
	assert.NoError(s.T(), err)

	userInput := model.CreateUserInput{
		ID:       userID,
		Username: userName.String(),
		Email:    email.String(),
	}
	err = s.service.Create(ctx, userInput)
	assert.NoError(s.T(), err)

	user, err := s.userRepo.FindByID(ctx, userID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), userID, user.ID())
	assert.Equal(s.T(), userName.String(), user.Username().String())
	assert.Equal(s.T(), email.String(), user.Email().String())
}

func (s *UserServiceTestSuite) TestCreateUserWithExistingUsername() {
	ctx := s.ctx
	userID := uuid.New()
	userName, err := userDomain.NewUsername("Test user")
	assert.NoError(s.T(), err)
	email, err := userDomain.NewEmail("test@example.com")
	assert.NoError(s.T(), err)

	userInput := model.CreateUserInput{
		ID:       userID,
		Username: userName.String(),
		Email:    email.String(),
	}
	err = s.service.Create(ctx, userInput)
	assert.NoError(s.T(), err)

	// Attempt to create another user with the same username
	userID2 := uuid.New()
	email2, err := userDomain.NewEmail("test2@example.com")
	assert.NoError(s.T(), err)

	userInput2 := model.CreateUserInput{
		ID:       userID2,
		Username: userName.String(),
		Email:    email2.String(),
	}
	err = s.service.Create(ctx, userInput2)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUsernameAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestCreateUserWithExistingEmail() {
	ctx := s.ctx
	userID := uuid.New()
	userName, err := userDomain.NewUsername("Test user")
	assert.NoError(s.T(), err)
	email, err := userDomain.NewEmail("test@example.com")
	assert.NoError(s.T(), err)

	userInput := model.CreateUserInput{
		ID:       userID,
		Username: userName.String(),
		Email:    email.String(),
	}
	err = s.service.Create(ctx, userInput)
	assert.NoError(s.T(), err)

	// Attempt to create another user with the same email
	userID2 := uuid.New()
	userName2, err := userDomain.NewUsername("Test user 2")
	assert.NoError(s.T(), err)

	userInput2 := model.CreateUserInput{
		ID:       userID2,
		Username: userName2.String(),
		Email:    email.String(),
	}
	err = s.service.Create(ctx, userInput2)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrEmailAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestGetUserByID() {
	ctx := s.ctx
	userID := uuid.New()
	userName, err := userDomain.NewUsername("Test user")
	assert.NoError(s.T(), err)
	email, err := userDomain.NewEmail("test@example.com")
	assert.NoError(s.T(), err)

	userInput := model.CreateUserInput{
		ID:       userID,
		Username: userName.String(),
		Email:    email.String(),
	}
	err = s.service.Create(ctx, userInput)
	assert.NoError(s.T(), err)

	// Test code for getting a user by ID
	user, err := s.service.GetByID(ctx, userID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), userID, user.ID)
	assert.Equal(s.T(), userName.String(), user.Username)
	assert.Equal(s.T(), email.String(), user.Email)
}

func (s *UserServiceTestSuite) TestUpdateUser() {
	// Test code for updating a user
}

func (s *UserServiceTestSuite) TestDeleteUser() {
	// Test code for deleting a user
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
