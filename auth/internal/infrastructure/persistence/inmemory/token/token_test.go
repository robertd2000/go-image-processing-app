package tokenmem_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	inmemory "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
)

var refreshTTL = 1 * time.Minute

func newRepo() (tokenDomain.TokenRepository, context.Context) {
	return inmemory.NewTokenRepository(), context.Background()
}

func TestTokenRepository_SaveAndIsValid(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Create(ctx, userID, token, expiresAt))
}

func TestTokenRepository_GetByToken(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Create(ctx, userID, token, expiresAt))

	t.Run("success", func(t *testing.T) {
		got, err := repo.GetByHash(ctx, token)

		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, userID, got.UserID())
	})

	t.Run("not found", func(t *testing.T) {
		got, err := repo.GetByHash(ctx, "unknown")

		require.Error(t, err)
		require.Nil(t, got)
	})
}

func TestTokenRepository_Update(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	oldToken := "old"
	newToken := "new"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Create(ctx, userID, oldToken, expiresAt))

	require.NoError(t, repo.Update(ctx, userID, oldToken, newToken))
}

func TestTokenRepository_Revoke(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Create(ctx, userID, token, expiresAt))
	require.NoError(t, repo.Revoke(ctx, token))
}

func TestTokenRepository_RevokeByToken(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Create(ctx, userID, token, expiresAt))
	require.NoError(t, repo.Revoke(ctx, token))
}
