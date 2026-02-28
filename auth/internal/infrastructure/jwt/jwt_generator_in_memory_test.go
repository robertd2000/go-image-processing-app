package jwt_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
)

func TestInMemoryTokenGenerator_GenerateAccess(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		userID  uuid.UUID
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			userID:  userID,
			want:    "access_" + userID.String(),
			wantErr: false,
		},
		{
			name:    "generate error",
			userID:  userID,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := jwt.NewInMemoryTokenGenerator()

			if tt.wantErr {
				g.GenerateErr = assertTestError()
			}

			got, err := g.GenerateAccess(tt.userID)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("GenerateAccess() unexpected error: %v", err)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("GenerateAccess() expected error, got nil")
			}

			if got != tt.want {
				t.Errorf("GenerateAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func assertTestError() error {
	return &testError{}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
