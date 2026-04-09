package jwt_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
)

func TestInMemoryTokenGenerator_GenerateAccess(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
		{
			name:    "generate error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := jwt.NewInMemoryTokenGenerator()

			if tt.wantErr {
				g.GenerateErr = assertTestError()
			}

			got, err := g.GenerateAccess(model.ClaimsInput{
				UserID: userID,
				Roles:  []string{"user"},
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == "" {
				t.Fatal("expected non-empty token")
			}

			if !strings.HasPrefix(got, "access_") {
				t.Fatalf("unexpected token format: %s", got)
			}

			claims, err := g.ValidateAccess(got)
			if err != nil {
				t.Fatalf("validate failed: %v", err)
			}

			if claims.UserID != userID {
				t.Fatalf("userID mismatch: got %v want %v", claims.UserID, userID)
			}
		})
	}
}

func (e *testError) Error() string { return "test error" }

func assertTestError() error {
	return &testError{}
}

type testError struct{}
