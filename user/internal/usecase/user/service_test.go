package user_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	outboxmem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/outbox"
	txmanagermem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/txmanager"
	usermem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/port"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserService interface {
	Create(ctx context.Context, userInput model.CreateUserInput) error
	CreateFromEvent(ctx context.Context, input model.CreateUserInput) error
	Update(ctx context.Context, input model.UpdateUserInput) error
	UpdateProfile(ctx context.Context, input model.UpdateProfileInput) error
	UpdateSettings(ctx context.Context, input model.UpdateSettingsInput) error
	Delete(ctx context.Context, userID uuid.UUID) error
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error)
	GetByEmail(ctx context.Context, email string) (*model.UserOutput, error)
	List(ctx context.Context, filter model.UserFilterInput) ([]*model.UserOutput, error)
	Count(ctx context.Context, filter model.UserFilterInput) (int, error)
	Ban(ctx context.Context, userID uuid.UUID, reason string) error
	Restore(ctx context.Context, userID uuid.UUID) error
}

type UserServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service UserService

	userRepo   userDomain.UserRepository
	outboxRepo port.OutboxRepository

	tx port.TxManager
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.userRepo = usermem.NewUserRepository()
	s.outboxRepo = outboxmem.NewRepository()

	s.tx = &txmanagermem.FakeTxManager{}

	s.service = user.NewUserService(s.userRepo, s.outboxRepo, s.tx)
}

// CreateUser
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

// CreateUserFromEvent
func (s *UserServiceTestSuite) TestCreateUserFromEvent() {
	input := s.newCreateUserInput()
	err := s.service.CreateFromEvent(s.ctx, input)
	assert.NoError(s.T(), err)

	user := s.mustGetUserFromRepo(input.ID)

	assert.Equal(s.T(), input.ID, user.ID())
	assert.Equal(s.T(), input.Username, user.Username().String())
	assert.Equal(s.T(), input.Email, user.Email().String())
}

func (s *UserServiceTestSuite) TestCreateUserFromEventAlreadyExists() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.CreateFromEvent(s.ctx, input)

	assert.NoError(s.T(), err) // should not return error if user already exists
}

// GetUserByID
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

func (s *UserServiceTestSuite) TestGetDeletedUserByID() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	user, err := s.service.GetByID(s.ctx, input.ID)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestGetBannedUserByID() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "rules violation")
	assert.NoError(s.T(), err)

	user, err := s.service.GetByID(s.ctx, input.ID)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

// GetUserByEmail
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

func (s *UserServiceTestSuite) TestGetDeletedUserByEmail() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	user, err := s.service.GetByEmail(s.ctx, input.Email)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestGetBannedUserByEmail() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "rules violation")
	assert.NoError(s.T(), err)

	user, err := s.service.GetByEmail(s.ctx, input.Email)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

// UpdateUser
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

func (s *UserServiceTestSuite) TestUpdateDeletedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	update := model.UpdateUserInput{
		UserID:   input.ID,
		Username: strPtr("newusername"),
	}

	err = s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateBannedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "rules violation")
	assert.NoError(s.T(), err)

	update := model.UpdateUserInput{
		UserID:   input.ID,
		Username: strPtr("newusername"),
	}

	err = s.service.Update(s.ctx, update)

	assert.Error(s.T(), err)
}

// UpdateProfile
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

func (s *UserServiceTestSuite) TestUpdateProfileDeletedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	update := model.UpdateProfileInput{
		UserID: input.ID,
		Bio:    strPtr("hello world"),
	}

	err = s.service.UpdateProfile(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateProfileBannedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "rules violation")
	assert.NoError(s.T(), err)

	update := model.UpdateProfileInput{
		UserID: input.ID,
		Bio:    strPtr("hello world"),
	}

	err = s.service.UpdateProfile(s.ctx, update)

	assert.Error(s.T(), err)
}

// UpdateSettings
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

func (s *UserServiceTestSuite) TestUpdateSettingsDeletedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	update := model.UpdateSettingsInput{
		UserID:   input.ID,
		IsPublic: boolPtr(false),
	}

	err = s.service.UpdateSettings(s.ctx, update)

	assert.Error(s.T(), err)
	assert.Equal(s.T(), err, userDomain.ErrUserNotFound)
}

func (s *UserServiceTestSuite) TestUpdateSettingsBannedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "rules violation")
	assert.NoError(s.T(), err)

	update := model.UpdateSettingsInput{
		UserID:   input.ID,
		IsPublic: boolPtr(false),
	}

	err = s.service.UpdateSettings(s.ctx, update)

	assert.Error(s.T(), err)
}

// DeleteUser
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

// BanUser
func (s *UserServiceTestSuite) TestBanUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "violate rules")
	assert.NoError(s.T(), err)

	user, err := s.userRepo.FindByID(s.ctx, input.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), user.Status(), userDomain.StatusBanned)
}

func (s *UserServiceTestSuite) TestBanDeletedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	err = s.service.Ban(s.ctx, input.ID, "violate rules")
	assert.Error(s.T(), err)

	user, err := s.service.GetByID(s.ctx, input.ID)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
}

func (s *UserServiceTestSuite) TestBanUserNotFound() {
	nonExistentID := uuid.New()
	err := s.service.Ban(s.ctx, nonExistentID, "violate rules")
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

// RestoreUser
func (s *UserServiceTestSuite) TestRestoreUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Delete(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	err = s.service.Restore(s.ctx, input.ID)
	assert.NoError(s.T(), err)

	user, err := s.userRepo.FindByID(s.ctx, input.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), user.Status(), userDomain.StatusActive)
}

func (s *UserServiceTestSuite) TestRestoreBannedUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Ban(s.ctx, input.ID, "violate rules")
	assert.NoError(s.T(), err)

	err = s.service.Restore(s.ctx, input.ID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)

	user, err := s.userRepo.FindByID(s.ctx, input.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), user.Status(), userDomain.StatusBanned)
}

func (s *UserServiceTestSuite) TestRestoreActiveUser() {
	input := s.newCreateUserInput()
	s.createUser(input)

	err := s.service.Restore(s.ctx, input.ID)
	assert.Error(s.T(), err)

	user, err := s.service.GetByID(s.ctx, input.ID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
}

func (s *UserServiceTestSuite) TestRestoreUserNotFound() {
	nonExistentID := uuid.New()
	err := s.service.Restore(s.ctx, nonExistentID)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), userDomain.ErrUserNotFound, err)
}

// ListUsers
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

func (s *UserServiceTestSuite) Test_List_ReturnsOnlyActiveUsers() {
	// active
	s.createUser(model.CreateUserInput{
		ID:       uuid.New(),
		Username: "active",
		Email:    "active@test.com",
	})

	// banned
	s.createUser(model.CreateUserInput{
		ID:       uuid.New(),
		Username: "banned",
		Email:    "banned@test.com",
	})

	// deleted
	s.createUser(model.CreateUserInput{
		ID:       uuid.New(),
		Username: "deleted",
		Email:    "deleted@test.com",
	})

	filter, _ := userDomain.NewUserFilter(10, 0, nil, nil, "", "")
	users, err := s.userRepo.List(s.ctx, filter)
	s.Require().NoError(err)

	for _, u := range users {
		switch u.Username().String() {
		case "banned":
			u.UpdateStatus(userDomain.StatusBanned)
			_ = s.userRepo.Update(s.ctx, u)

		case "deleted":
			u.UpdateStatus(userDomain.StatusInactive)
			_ = s.userRepo.Update(s.ctx, u)
		}
	}

	result, err := s.service.List(s.ctx, model.UserFilterInput{})
	s.Require().NoError(err)

	s.Require().Len(result, 1)
	s.Equal("active", result[0].Username)
}

func (s *UserServiceTestSuite) Test_List_MapsToOutput() {
	input := model.CreateUserInput{
		ID:       uuid.New(),
		Username: "test",
		Email:    "test@test.com",
	}

	s.createUser(input)

	result, err := s.service.List(s.ctx, model.UserFilterInput{})
	s.Require().NoError(err)

	s.Require().Len(result, 1)

	user := result[0]
	s.Equal("test", user.Username)
	s.Equal("test@test.com", user.Email)
}

// CountUsers
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

// helpers
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

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func mustUsername(s string) userDomain.Username {
	u, err := userDomain.NewUsername(s)
	if err != nil {
		panic(err)
	}
	return u
}

func mustEmail(s string) userDomain.Email {
	e, err := userDomain.NewEmail(s)
	if err != nil {
		panic(err)
	}
	return e
}
