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

	// user := userDomain.NewUser(
	// 	userID,
	// 	userName,
	// 	email,
	// )
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

func (s *UserServiceTestSuite) TestGetUserByID() {
	// Test code for getting a user by ID
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
