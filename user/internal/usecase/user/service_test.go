package user_test

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	usermem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserService interface {
	Create(ctx context.Context, userInput model.CreateUserInput) error
	Update(ctx context.Context, input model.UpdateUserInput) error
	UpdateProfile(ctx context.Context, input model.UpdateProfileInput) error
	UpdateSettings(ctx context.Context, input model.UpdateSettingsInput) error
	Delete(ctx context.Context, userID uuid.UUID) error
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error)
	GetByEmail(ctx context.Context, email string) (*model.UserOutput, error)
	List(ctx context.Context, filter model.UserFilterInput) ([]*model.UserOutput, error)
	Count(ctx context.Context, filter model.UserFilterInput) (int, error)
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

func (s *UserServiceTestSuite) TestUpdateUser_IgnoreNilFields() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateUserInput{
		UserID: input.ID,
	}

	err := s.service.Update(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), input.Username, user.Username().String())
	assert.Equal(s.T(), input.Email, user.Email().String())
}

func (s *UserServiceTestSuite) TestUpdateUser_InvalidEmail() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateUserInput{
		UserID: input.ID,
		Email:  strPtr("invalid-email"),
	}

	err := s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateUser_DuplicateUsername() {
	user1 := s.newCreateUserInput()
	user2 := s.newCreateUserInputWith("user2", "user2@test.com")

	s.createUser(user1)
	s.createUser(user2)

	update := model.UpdateUserInput{
		UserID:   user2.ID,
		Username: strPtr(user1.Username),
	}

	err := s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUsernameAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestUpdateUser_DuplicateEmail() {
	user1 := s.newCreateUserInput()
	user2 := s.newCreateUserInputWith("user2", "user2@test.com")

	s.createUser(user1)
	s.createUser(user2)

	update := model.UpdateUserInput{
		UserID: user2.ID,
		Email:  strPtr(user1.Email),
	}

	err := s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrEmailAlreadyExists, err)
}

func (s *UserServiceTestSuite) TestUpdateUser_NotFound() {
	update := model.UpdateUserInput{
		UserID:   uuid.New(),
		Username: strPtr("newname"),
	}

	err := s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_Bio() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateProfileInput{
		UserID: input.ID,
		Bio:    strPtr("hello world"),
	}

	err := s.service.UpdateProfile(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), "hello world", *user.Profile().Bio())
}

func (s *UserServiceTestSuite) TestUpdateProfile_MultipleFields() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateProfileInput{
		UserID:   input.ID,
		Bio:      strPtr("bio"),
		Location: strPtr("Berlin"),
	}

	err := s.service.UpdateProfile(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), "bio", *user.Profile().Bio())
	assert.Equal(s.T(), "Berlin", *user.Profile().Location())
}

func (s *UserServiceTestSuite) TestUpdateProfile_IgnoreNil() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateProfileInput{
		UserID: input.ID,
	}

	err := s.service.UpdateProfile(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Nil(s.T(), user.Profile().Bio())
	assert.Nil(s.T(), user.Profile().Location())
}

func (s *UserServiceTestSuite) TestUpdateProfile_NotFound() {
	update := model.UpdateProfileInput{
		UserID: uuid.New(),
		Bio:    strPtr("bio"),
	}

	err := s.service.UpdateProfile(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestUpdateProfile_ClearBio() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateProfileInput{
		UserID: input.ID,
		Bio:    strPtr(""),
	}

	err := s.service.UpdateProfile(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), "", *user.Profile().Bio())
}

func (s *UserServiceTestSuite) TestUpdateSettings_IsPublic() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateSettingsInput{
		UserID:   input.ID,
		IsPublic: boolPtr(false),
	}

	err := s.service.UpdateSettings(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), false, user.Settings().IsPublic())
}

func (s *UserServiceTestSuite) TestUpdateSettings_Multiple() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateSettingsInput{
		UserID:             input.ID,
		IsPublic:           boolPtr(false),
		AllowNotifications: boolPtr(false),
		Theme:              strPtr("dark"),
	}

	err := s.service.UpdateSettings(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), false, user.Settings().IsPublic())
	assert.Equal(s.T(), false, user.Settings().AllowNotifications())
	assert.Equal(s.T(), "dark", user.Settings().Theme())
}

func (s *UserServiceTestSuite) TestUpdateSettings_IgnoreNil() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateSettingsInput{
		UserID: input.ID,
	}

	err := s.service.UpdateSettings(s.ctx, update)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), true, user.Settings().IsPublic()) // default
	assert.Equal(s.T(), "light", user.Settings().Theme())
}

func (s *UserServiceTestSuite) TestUpdateSettings_InvalidTheme() {
	input := s.newCreateUserInput()
	s.createUser(input)

	update := model.UpdateSettingsInput{
		UserID: input.ID,
		Theme:  strPtr("blue"),
	}

	err := s.service.UpdateSettings(s.ctx, update)

	assert.Error(s.T(), err)
}

func (s *UserServiceTestSuite) TestUpdateSettings_NotFound() {
	update := model.UpdateSettingsInput{
		UserID: uuid.New(),
	}

	err := s.service.UpdateSettings(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestDeleteUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	user, err := s.service.GetByID(s.ctx, input.ID)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestDeleteUserNotFound() {
	nonExistentID := uuid.New()
	err := s.service.Delete(s.ctx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestDeleteUserInvalidID() {
	invalidID := uuid.Nil
	err := s.service.Delete(s.ctx, invalidID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestDeleteUserAlreadyDeleted() {
	input := s.newCreateUserInput()
	s.createUser(input)
	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)
	err = s.service.Delete(s.ctx, input.ID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

func (s *UserServiceTestSuite) TestListUsers() {
	input := s.newCreateUserInput()
	s.createUser(input)

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  10,
		Offset: 0,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 1)
	assert.Equal(s.T(), input.ID, users[0].ID)
}

func (s *UserServiceTestSuite) TestListUsersWithSearch() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	user2 := s.newCreateUserInputWith("bob", "bob@example.com")
	s.createUser(user2)

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  10,
		Offset: 0,
		Search: "alice",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 1)
	assert.Equal(s.T(), user1.ID, users[0].ID)
}

func (s *UserServiceTestSuite) TestListUsersWithPagination() {
	for i := range 5 {
		input := s.newCreateUserInputWith(
			"user"+strconv.Itoa(i),
			"user"+strconv.Itoa(i)+"@example.com",
		)
		s.createUser(input)
	}

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  2,
		Offset: 1,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)
}

func (s *UserServiceTestSuite) TestListUsersWithSearchNoResults() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  10,
		Offset: 0,
		Search: "bob",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)
}

func (s *UserServiceTestSuite) TestListUsersWithPaginationBeyondRange() {
	for i := range 3 {
		input := s.newCreateUserInputWith(
			"user"+strconv.Itoa(i),
			"user"+strconv.Itoa(i)+"@example.com",
		)
		s.createUser(input)
	}

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  2,
		Offset: 5,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)
}

func (s *UserServiceTestSuite) TestListUsersWithInvalidPagination() {
	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  -1,
		Offset: -1,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)
}

func (s *UserServiceTestSuite) TestListUsersWithInvalidSearch() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	users, err := s.service.List(s.ctx, model.UserFilterInput{
		Limit:  10,
		Offset: 0,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 1)
}

func (s *UserServiceTestSuite) TestCountUsers() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	user2 := s.newCreateUserInputWith("bob", "bob@example.com")
	s.createUser(user2)

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithSearch() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	user2 := s.newCreateUserInputWith("bob", "bob@example.com")
	s.createUser(user2)

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Search: "alice",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithSearchNoResults() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Search: "bob",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithInvalidSearch() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithNoUsers() {
	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithInvalidPagination() {
	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Limit:  -1,
		Offset: -1,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithPagination() {
	for i := range 5 {
		input := s.newCreateUserInputWith(
			"user"+strconv.Itoa(i),
			"user"+strconv.Itoa(i)+"@example.com",
		)
		s.createUser(input)
	}

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Limit:  2,
		Offset: 0,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 5, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithPaginationBeyondRange() {
	for i := range 3 {
		input := s.newCreateUserInputWith(
			"user"+strconv.Itoa(i),
			"user"+strconv.Itoa(i)+"@example.com",
		)
		s.createUser(input)
	}

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Limit:  2,
		Offset: 5,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 3, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithInvalidSearchAndPagination() {
	user1 := s.newCreateUserInputWith("alice", "alice@example.com")
	s.createUser(user1)

	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Limit:  -1,
		Offset: -1,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
}

func (s *UserServiceTestSuite) TestCountUsersWithInvalidSearchAndPaginationNoUsers() {
	count, err := s.service.Count(s.ctx, model.UserFilterInput{
		Limit:  -1,
		Offset: -1,
		Search: "",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
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
