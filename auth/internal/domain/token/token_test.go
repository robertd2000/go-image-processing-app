package token_test

import (
	"testing"

	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
)

func TestNewAccessRefresh(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		accessToken  string
		refreshToken string
		want         *token.Tokens
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := token.NewTokens(tt.accessToken, tt.refreshToken)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewAccessRefresh() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewAccessRefresh() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewAccessRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}
