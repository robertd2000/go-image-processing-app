package usermem_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

// ---------- helpers ----------

func newRepo() userDomain.UserRepository {
	return usermem.NewUserRepository()
}

func generateTestUserData(t *testing.T, email, username, firstname, lastname string) *userDomain.User {
	t.Helper()

	userID := uuid.New()
	userPassword := "hashedpassword"

	user, err := userDomain.NewUser(userID, username, firstname, lastname, &email, userPassword)
	require.NoError(t, err)

	return user
}

func createTestUser(
	t *testing.T,
	repo userDomain.UserRepository,
	email, username, firstname, lastname string,
) *userDomain.User {
	t.Helper()

	user := generateTestUserData(t, email, username, firstname, lastname)
	require.NoError(t, repo.Create(ctx, user))
	return user
}

// ---------- tests ----------

func TestUserRepository_Create(t *testing.T) {
	repo := newRepo()

	user := generateTestUserData(t, "test@example.com", "test 1", "test", "1")

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	t.Run("already exists", func(t *testing.T) {
		err := repo.Create(ctx, user)
		assert.Error(t, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	repo := newRepo()

	user := createTestUser(t, repo, "test@example.com", "test 1", "test", "1")

	updated := user.Clone()
	updated.UpdateEmail("test-update@example.com")

	require.NoError(t, repo.Update(ctx, updated))

	got, err := repo.GetByID(ctx, user.ID())
	require.NoError(t, err)

	assert.Equal(t, updated.Email(), got.Email())
}

func TestUserRepository_Delete(t *testing.T) {
	repo := newRepo()

	user := createTestUser(t, repo, "test@example.com", "test 1", "test", "1")

	require.NoError(t, repo.Delete(ctx, user.ID()))

	got, err := repo.GetByID(ctx, user.ID())
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo := newRepo()

	user := createTestUser(t, repo, "test@example.com", "test 1", "test", "1")

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, user.Email())
		require.NoError(t, err)
		assert.Equal(t, user, found)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "missing@example.com")
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	repo := newRepo()

	user := createTestUser(t, repo, "test@example.com", "test 1", "test", "1")

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByUsername(ctx, user.Username())
		require.NoError(t, err)
		assert.Equal(t, user, found)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByUsername(ctx, "missing")
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	repo := newRepo()

	user := createTestUser(t, repo, "test@example.com", "test 1", "test", "1")

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID())
		require.NoError(t, err)
		assert.Equal(t, user, found)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}
