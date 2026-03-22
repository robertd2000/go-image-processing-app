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

	require.NoError(t, repo.Save(ctx, userID, token, expiresAt))

	ok, err := repo.IsValid(ctx, userID, token)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestTokenRepository_IsValid(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Save(ctx, userID, token, expiresAt))

	tests := []struct {
		name   string
		userID uuid.UUID
		token  string
		valid  bool
	}{
		{
			name:   "valid token",
			userID: userID,
			token:  token,
			valid:  true,
		},
		{
			name:   "unknown token",
			userID: userID,
			token:  "unknown",
			valid:  false,
		},
		{
			name:   "wrong user",
			userID: uuid.New(),
			token:  token,
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := repo.IsValid(ctx, tt.userID, tt.token)
			require.NoError(t, err)
			require.Equal(t, tt.valid, ok)
		})
	}
}

func TestTokenRepository_GetByToken(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Save(ctx, userID, token, expiresAt))

	t.Run("success", func(t *testing.T) {
		got, err := repo.GetByToken(ctx, token)

		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, userID, got.UserID())
	})

	t.Run("not found", func(t *testing.T) {
		got, err := repo.GetByToken(ctx, "unknown")

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

	require.NoError(t, repo.Save(ctx, userID, oldToken, expiresAt))

	require.NoError(t, repo.Update(ctx, userID, oldToken, newToken))

	ok, err := repo.IsValid(ctx, userID, oldToken)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = repo.IsValid(ctx, userID, newToken)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestTokenRepository_Revoke(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Save(ctx, userID, token, expiresAt))
	require.NoError(t, repo.Revoke(ctx, token))

	ok, err := repo.IsValid(ctx, userID, token)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestTokenRepository_RevokeByToken(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	token := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	require.NoError(t, repo.Save(ctx, userID, token, expiresAt))
	require.NoError(t, repo.Revoke(ctx, token))

	ok, err := repo.IsValid(ctx, userID, token)
	require.NoError(t, err)
	require.False(t, ok)
}
