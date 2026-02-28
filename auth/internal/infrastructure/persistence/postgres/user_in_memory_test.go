package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserSuccess(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	userEmail := "test@example.com"
	username := "test 1"
	userFirstname := "test"
	userLastname := "1"

	user := generateTestUserData(t, userEmail, username, userFirstname, userLastname)

	err := repo.Create(context.Background(), user)

	assert.NoError(t, err)
}

func TestCreateUserErrAlreadyExists(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	userEmail := "test@example.com"
	username := "test 1"
	userFirstname := "test"
	userLastname := "1"

	user := generateTestUserData(t, userEmail, username, userFirstname, userLastname)

	err := repo.Create(context.Background(), user)

	assert.NoError(t, err)

	err = repo.Create(context.Background(), user)

	assert.Error(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	user := createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByEmail(context.Background(), user.Email())
	assert.NoError(t, err)
	assert.Equal(t, user, found)
}

func TestGetUserByEmailNotFound(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByEmail(context.Background(), "example@example.com")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestGetUserByUsername(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	user := createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByUsername(context.Background(), user.Username())
	assert.NoError(t, err)
	assert.Equal(t, user, found)
}

func TestGetUserByUsernameNotFound(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByUsername(context.Background(), "example@example.com")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestGetUserByID(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	user := createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByID(context.Background(), user.ID())
	assert.NoError(t, err)
	assert.Equal(t, user, found)
}

func TestGetUserByIDNotFound(t *testing.T) {
	repo := postgres.NewUserInMemoryRepository()

	email := "test@example.com"
	username := "test 1"
	firstname := "test"
	lastname := "1"

	createTestUser(t, repo, email, username, firstname, lastname)

	found, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, found)
}

func generateTestUserData(t *testing.T, email, username, firstname, lastname string) *userDomain.User {
	userID := uuid.New()
	userPassword := "hashedpassword"
	user, err := userDomain.NewUser(userID, username, firstname, lastname, &email, userPassword)
	assert.NoError(t, err)

	return user
}

func createTestUser(t *testing.T, repo userDomain.UserRepository, email, username, firstname, lastname string) *userDomain.User {
	user := generateTestUserData(t, email, username, firstname, lastname)
	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	return user
}
