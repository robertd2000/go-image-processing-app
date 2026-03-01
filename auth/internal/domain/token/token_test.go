package token_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
)

func TestNewTokens(t *testing.T) {
	now := time.Now()
	expires := now.Add(time.Hour)
	userID := uuid.New()

	tests := []struct {
		name         string
		userID       uuid.UUID
		accessToken  string
		refreshToken string
		expiresAt    time.Time
		wantErr      bool
	}{
		{
			name:         "valid tokens",
			userID:       userID,
			accessToken:  "access123",
			refreshToken: "refresh123",
			expiresAt:    expires,
			wantErr:      false,
		},
		{
			name:         "empty access token",
			userID:       userID,
			accessToken:  "",
			refreshToken: "refresh123",
			expiresAt:    expires,
			wantErr:      true,
		},
		{
			name:         "empty refresh token",
			userID:       userID,
			accessToken:  "access123",
			refreshToken: "",
			expiresAt:    expires,
			wantErr:      true,
		},
		{
			name:         "both tokens empty",
			userID:       userID,
			accessToken:  "",
			refreshToken: "",
			expiresAt:    expires,
			wantErr:      true,
		},
		{
			name:         "zero user id",
			userID:       uuid.Nil,
			accessToken:  "access123",
			refreshToken: "refresh123",
			expiresAt:    expires,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := token.NewTokens(
				tt.userID,
				tt.accessToken,
				tt.refreshToken,
				tt.expiresAt,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.accessToken, got.AccessToken())
			assert.Equal(t, tt.refreshToken, got.RefreshToken())

			assert.False(t, got.IsRevoked())
			assert.False(t, got.IsExpired(now))
		})
	}
}
