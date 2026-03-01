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

func newValidTokens(t *testing.T, userID uuid.UUID, access string, exp time.Time) *tokenDomain.Tokens {
	t.Helper()

	tokens, err := tokenDomain.NewTokens(
		userID,
		access,
		"refresh-"+access,
		exp,
	)
	require.NoError(t, err)
	return tokens
}

func TestTokenRepository_Save_And_IsValid(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()
	tokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, tokens.AccessToken()))

	ok, err := repo.IsValid(ctx, userID, tokens.AccessToken())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestTokenRepository_IsValid_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	ok, err := repo.IsValid(ctx, uuid.New(), "unknown")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestTokenRepository_IsValid_WrongUser(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()
	otherUser := uuid.New()

	tokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, tokens.AccessToken()))

	ok, err := repo.IsValid(ctx, otherUser, tokens.AccessToken())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestTokenRepository_GetByToken(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()
	tokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, tokens.AccessToken()))

	got, err := repo.GetByToken(ctx, tokens.AccessToken())
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, userID, got.UserID())
	require.Equal(t, tokens.AccessToken(), got.AccessToken())
}

func TestTokenRepository_GetByToken_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	got, err := repo.GetByToken(ctx, "unknown")
	require.Error(t, err)
	require.Nil(t, got)
}

func TestTokenRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()

	oldTokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))
	newTokens := newValidTokens(t, userID, "access2", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, oldTokens.AccessToken()))

	require.NoError(t, repo.Update(
		ctx,
		userID,
		oldTokens.AccessToken(),
		newTokens.AccessToken(),
	))

	ok, err := repo.IsValid(ctx, userID, oldTokens.AccessToken())
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = repo.IsValid(ctx, userID, newTokens.AccessToken())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestTokenRepository_Revoke(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()
	tokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, tokens.AccessToken()))
	require.NoError(t, repo.Revoke(ctx, userID, tokens.AccessToken()))

	ok, err := repo.IsValid(ctx, userID, tokens.AccessToken())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestTokenRepository_RevokeByToken(t *testing.T) {
	ctx := context.Background()
	repo := inmemory.NewTokenRepository()

	userID := uuid.New()
	tokens := newValidTokens(t, userID, "access1", time.Now().Add(time.Hour))

	require.NoError(t, repo.Save(ctx, userID, tokens.AccessToken()))
	require.NoError(t, repo.RevokeByToken(ctx, tokens.AccessToken()))

	ok, err := repo.IsValid(ctx, userID, tokens.AccessToken())
	require.NoError(t, err)
	require.False(t, ok)
}
