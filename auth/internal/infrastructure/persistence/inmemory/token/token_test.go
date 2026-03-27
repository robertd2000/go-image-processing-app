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

func TestTokenRepository_CreateAndGet(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	familyID := uuid.New()

	tokenHash := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	token, err := tokenDomain.NewTokens(userID, tokenHash, expiresAt, familyID, uuid.Nil)
	require.NoError(t, err)

	require.NoError(t, repo.Create(ctx, token, 5))

	got, err := repo.GetByHash(ctx, tokenHash)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, userID, got.UserID())
	require.Equal(t, familyID, got.FamilyID())
	require.True(t, got.ExpiresAt().After(time.Now()))
}

func TestTokenRepository_GetByHash_NotFound(t *testing.T) {
	repo, ctx := newRepo()

	got, err := repo.GetByHash(ctx, "unknown")

	require.Error(t, err)
	require.Nil(t, got)
}

func TestTokenRepository_Revoke(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	familyID := uuid.New()

	tokenHash := "token1"
	expiresAt := time.Now().Add(refreshTTL)

	token, err := tokenDomain.NewTokens(userID, tokenHash, expiresAt, familyID, uuid.Nil)
	require.NoError(t, err)

	require.NoError(t, repo.Create(ctx, token, 5))

	// revoke
	require.NoError(t, repo.Revoke(ctx, token.ID()))

	// check revoked
	got, err := repo.GetByHash(ctx, tokenHash)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.True(t, got.IsRevoked())
}

func TestTokenRepository_Rotate(t *testing.T) {
	repo, ctx := newRepo()

	userID := uuid.New()
	familyID := uuid.New()

	oldHash := "old_token"
	newHash := "new_token"

	expiresAt := time.Now().Add(refreshTTL)

	oldToken, err := tokenDomain.NewTokens(userID, oldHash, expiresAt, familyID, uuid.Nil)
	require.NoError(t, err)

	require.NoError(t, repo.Create(ctx, oldToken, 5))

	parentID := oldToken.ID()

	newToken, err := tokenDomain.NewTokens(
		userID,
		newHash,
		expiresAt,
		familyID,
		parentID,
	)
	require.NoError(t, err)

	// rotate
	rotated, err := repo.Rotate(ctx, oldToken, newToken)
	require.NoError(t, err)
	require.False(t, rotated)

	// old token should be revoked
	oldFromDB, err := repo.GetByHash(ctx, oldHash)
	require.NoError(t, err)
	require.True(t, oldFromDB.IsRevoked())

	// new token should exist
	newFromDB, err := repo.GetByHash(ctx, newHash)
	require.NoError(t, err)
	require.NotNil(t, newFromDB)

	require.Equal(t, parentID, newFromDB.ParentID())
	require.Equal(t, familyID, newFromDB.FamilyID())
}
