package usermem_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
	usermem "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/inmemory/user"
)

type UserRepoTestSuite struct {
	suite.Suite

	ctx  context.Context
	repo userDomain.UserRepository
}

func (s *UserRepoTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.repo = usermem.NewUserRepository()
}

//
// =====================
// HELPERS
// =====================
//

func (s *UserRepoTestSuite) newUser(username, email string) *userDomain.User {
	id := uuid.New()

	uName, err := userDomain.NewUsername(username)
	s.Require().NoError(err)

	uEmail, err := userDomain.NewEmail(email)
	s.Require().NoError(err)

	return userDomain.NewUser(id, uName, uEmail)
}

func (s *UserRepoTestSuite) createUser(username, email string) *userDomain.User {
	user := s.newUser(username, email)

	err := s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	return user
}

//
// =====================
// TESTS
// =====================
//

func (s *UserRepoTestSuite) TestCreateAndFindByID() {
	user := s.createUser("test", "test@test.com")

	found, err := s.repo.FindByID(s.ctx, user.ID())

	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID(), found.ID())
}

func (s *UserRepoTestSuite) TestFindByID_NotFound() {
	_, err := s.repo.FindByID(s.ctx, uuid.New())

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

func (s *UserRepoTestSuite) TestFindByEmail() {
	user := s.createUser("test", "test@test.com")

	email, _ := userDomain.NewEmail("test@test.com")

	found, err := s.repo.FindByEmail(s.ctx, email)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID(), found.ID())
}

func (s *UserRepoTestSuite) TestFindByEmail_NotFound() {
	email, _ := userDomain.NewEmail("none@test.com")

	_, err := s.repo.FindByEmail(s.ctx, email)

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

func (s *UserRepoTestSuite) TestFindByUsername() {
	user := s.createUser("test", "test@test.com")

	username, _ := userDomain.NewUsername("test")

	found, err := s.repo.FindByUsername(s.ctx, username)

	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID(), found.ID())
}

func (s *UserRepoTestSuite) TestFindByUsername_NotFound() {
	username, _ := userDomain.NewUsername("none")

	_, err := s.repo.FindByUsername(s.ctx, username)

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

//
// =====================
// DELETE (SOFT DELETE)
// =====================
//

func (s *UserRepoTestSuite) TestDelete_UserBecomesInactive() {
	user := s.createUser("test", "test@test.com")

	err := s.repo.Delete(s.ctx, user.ID())
	s.NoError(err)

	email, _ := userDomain.NewEmail("test@test.com")

	_, err = s.repo.FindByEmail(s.ctx, email)

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

func (s *UserRepoTestSuite) TestDelete_FindByIDStillReturnsUser() {
	user := s.createUser("test", "test@test.com")

	err := s.repo.Delete(s.ctx, user.ID())
	s.NoError(err)

	found, err := s.repo.FindByID(s.ctx, user.ID())

	s.NoError(err)
	s.NotNil(found)
	s.Equal(userDomain.StatusInactive, found.Status())
}

//
// =====================
// EXISTS
// =====================
//

func (s *UserRepoTestSuite) TestExistsByEmail() {
	s.createUser("test", "test@test.com")

	email, _ := userDomain.NewEmail("test@test.com")

	exists, err := s.repo.ExistsByEmail(s.ctx, email)

	s.NoError(err)
	s.True(exists)
}

func (s *UserRepoTestSuite) TestExistsByUsername() {
	s.createUser("test", "test@test.com")

	username, _ := userDomain.NewUsername("test")

	exists, err := s.repo.ExistsByUsername(s.ctx, username)

	s.NoError(err)
	s.True(exists)
}

//
// =====================
// UPDATE
// =====================
//

func (s *UserRepoTestSuite) TestUpdate() {
	user := s.createUser("test", "test@test.com")

	newUsername, _ := userDomain.NewUsername("newname")
	err := user.ChangeUsername(newUsername)
	s.Require().NoError(err)

	err = s.repo.Update(s.ctx, user)
	s.NoError(err)

	found, err := s.repo.FindByID(s.ctx, user.ID())
	s.NoError(err)

	s.Equal("newname", found.Username().String())
}

func (s *UserRepoTestSuite) TestUpdate_NotFound() {
	user := s.newUser("test", "test@test.com")

	err := s.repo.Update(s.ctx, user)

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

//
// =====================
// DELETE NOT FOUND
// =====================
//

func (s *UserRepoTestSuite) TestDelete_NotFound() {
	err := s.repo.Delete(s.ctx, uuid.New())

	s.Error(err)
	s.Equal(userDomain.ErrUserNotFound, err)
}

// =====================
// LIST AND COUNT
// =====================
//

func (s *UserRepoTestSuite) TestList() {
	s.createUser("test1", "test1@test.com")
	s.createUser("test2", "test2@test.com")
	s.createUser("test3", "test3@test.com")

	filter := userDomain.UserFilter{
		Limit:  10,
		Offset: 0,
	}
	users, err := s.repo.List(s.ctx, filter)

	s.NoError(err)
	s.Len(users, 3)
}

func (s *UserRepoTestSuite) TestList_Empty() {
	filter := userDomain.UserFilter{
		Limit:  10,
		Offset: 0,
	}
	users, err := s.repo.List(s.ctx, filter)

	s.NoError(err)
	s.Len(users, 0)
}

func (s *UserRepoTestSuite) TestCount() {
	s.createUser("test1", "test1@test.com")
	s.createUser("test2", "test2@test.com")

	filter := userDomain.UserFilter{
		Limit:  10,
		Offset: 0,
	}
	count, err := s.repo.Count(s.ctx, filter)

	s.NoError(err)
	s.Equal(2, count)
}

func (s *UserRepoTestSuite) TestCount_Empty() {
	filter := userDomain.UserFilter{
		Limit:  10,
		Offset: 0,
	}
	count, err := s.repo.Count(s.ctx, filter)

	s.NoError(err)
	s.Equal(0, count)
}

// =====================
// RUN
// =====================
//

func TestUserRepoTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepoTestSuite))
}
