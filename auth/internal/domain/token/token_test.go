package token_test

import (
	"testing"

	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
)

func TestNewTokens(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		refreshToken string
		wantAccess   string
		wantRefresh  string
		wantErr      bool
	}{
		{
			name:         "valid tokens",
			accessToken:  "access123",
			refreshToken: "refresh123",
			wantAccess:   "access123",
			wantRefresh:  "refresh123",
			wantErr:      false,
		},
		{
			name:         "empty access token",
			accessToken:  "",
			refreshToken: "refresh123",
			wantErr:      true,
		},
		{
			name:         "empty refresh token",
			accessToken:  "access123",
			refreshToken: "",
			wantErr:      true,
		},
		{
			name:         "both tokens empty",
			accessToken:  "",
			refreshToken: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := token.NewTokens(tt.accessToken, tt.refreshToken)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTokens() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got == nil {
				t.Fatal("NewTokens() returned nil without error")
			}

			if got.GetAccessToken() != tt.wantAccess {
				t.Errorf("GetAccessToken() = %v, want %v",
					got.GetAccessToken(), tt.wantAccess)
			}

			if got.GetRefreshToken() != tt.wantRefresh {
				t.Errorf("GetRefreshToken() = %v, want %v",
					got.GetRefreshToken(), tt.wantRefresh)
			}
		})
	}
}
