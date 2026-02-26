package role_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
)

func TestRole_Name(t *testing.T) {
	tests := []struct {
		id          uuid.UUID
		name        role.Name
		permissions []role.Permission
		want        role.Name
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			r, err := role.New(tt.id, tt.name, tt.permissions)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := r.Name()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
