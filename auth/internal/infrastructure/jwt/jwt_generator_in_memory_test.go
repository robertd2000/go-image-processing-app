package jwt_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
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

			got, err := g.GenerateAccess(userID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// ✅ проверяем, что токен есть
			if got == "" {
				t.Fatal("expected non-empty token")
			}

			// ✅ проверяем префикс
			if !strings.HasPrefix(got, "access_") {
				t.Fatalf("unexpected token format: %s", got)
			}

			// ✅ проверяем, что он валидируется
			id, err := g.ValidateAccess(got)
			if err != nil {
				t.Fatalf("validate failed: %v", err)
			}

			if id != userID {
				t.Fatalf("userID mismatch: got %v want %v", id, userID)
			}
		})
	}
}

func (e *testError) Error() string { return "test error" }

func assertTestError() error {
	return &testError{}
}

type testError struct{}
