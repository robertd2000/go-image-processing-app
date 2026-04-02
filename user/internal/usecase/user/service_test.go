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
	Update(ctx context.Context, input model.UpdateUserInput) error
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error)
	GetByEmail(ctx context.Context, email string) (*model.UserOutput, error)
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
	input := s.newCreateUserInput()

	s.createUser(input)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), input.ID, user.ID())
	assert.Equal(s.T(), input.Username, user.Username().String())
	assert.Equal(s.T(), input.Email, user.Email().String())
}

func (s *UserServiceTestSuite) TestCreateUserWithExistingUsername() {
	input1 := s.newCreateUserInput()
	s.createUser(input1)

	input2 := s.newCreateUserInputWith(
		input1.Username,
		"test2@example.com",
	)

	err := s.service.Create(s.ctx, input2)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUsernameAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestCreateUserWithExistingEmail() {
	input1 := s.newCreateUserInput()
	s.createUser(input1)

	input2 := s.newCreateUserInputWith(
		"anotheruser",
		input1.Email,
	)

	err := s.service.Create(s.ctx, input2)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrEmailAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestGetUserByID() {
	input := s.newCreateUserInput()
	s.createUser(input)

	user, err := s.service.GetByID(s.ctx, input.ID)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)

	assert.Equal(s.T(), input.ID, user.ID)
	assert.Equal(s.T(), input.Username, user.Username)
	assert.Equal(s.T(), input.Email, user.Email)

	// profile defaults
	assert.Nil(s.T(), user.Profile.Bio)
	assert.Nil(s.T(), user.Profile.Location)
	assert.Nil(s.T(), user.Profile.Website)

	// settings defaults
	assert.True(s.T(), user.Settings.IsPublic)
	assert.Equal(s.T(), "light", user.Settings.Theme)
}

func (s *UserServiceTestSuite) TestGetUserByIDNotFound() {
	nonExistentID := uuid.New()
	user, err := s.service.GetByID(s.ctx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestGetUserByIDInvalidID() {
	invalidID := uuid.Nil
	user, err := s.service.GetByID(s.ctx, invalidID)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestGetUserByEmail() {
	input := s.newCreateUserInput()
	s.createUser(input)

	user, err := s.service.GetByEmail(s.ctx, input.Email)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), input.ID, user.ID)
	assert.Equal(s.T(), input.Username, user.Username)
	assert.Equal(s.T(), input.Email, user.Email)
}

func (s *UserServiceTestSuite) TestGetUserByEmailNotFound() {
	nonExistentEmail := "nonexistent@example.com"
	user, err := s.service.GetByEmail(s.ctx, nonExistentEmail)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateUser_Username() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateUserInput{
		UserID:   input.ID,
		Username: strPtr("newusername"),
	}

	err := s.service.Update(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), "newusername", user.Username().String())
	assert.Equal(s.T(), input.Email, user.Email().String()) // не изменился
}

func (s *UserServiceTestSuite) TestUpdateUser_MultipleFields() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateUserInput{
		UserID:    input.ID,
		Username:  strPtr("newusername"),
		FirstName: strPtr("John"),
		LastName:  strPtr("Doe"),
	}

	err := s.service.Update(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), "newusername", user.Username().String())
	assert.Equal(s.T(), "John", user.FirstName())
	assert.Equal(s.T(), "Doe", user.LastName())
}

func (s *UserServiceTestSuite) TestDeleteUser() {
	// Test code for deleting a user
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) newCreateUserInput() model.CreateUserInput {
	return model.CreateUserInput{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
}

func (s *UserServiceTestSuite) newCreateUserInputWith(
	username string,
	email string,
) model.CreateUserInput {
	return model.CreateUserInput{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
	}
}

func (s *UserServiceTestSuite) createUser(input model.CreateUserInput) {
	err := s.service.Create(s.ctx, input)
	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) mustGetUserFromRepo(id uuid.UUID) *userDomain.User {
	u, err := s.userRepo.FindByID(s.ctx, id)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), u)
	return u
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
