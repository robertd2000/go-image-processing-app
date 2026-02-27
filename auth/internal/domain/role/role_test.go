package role_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
)

func TestRole_Name(t *testing.T) {
	tests := []struct {
		name        string
		id          uuid.UUID
		roleName    role.Name
		permissions []role.Permission
		want        role.Name
		wantErr     bool
	}{
		{
			name:        "admin role",
			id:          uuid.New(),
			roleName:    role.Admin,
			permissions: []role.Permission{role.PermUserRead, role.PermUserWrite},
			want:        role.Admin,
			wantErr:     false,
		},
		{
			name:        "user role",
			id:          uuid.New(),
			roleName:    role.User,
			permissions: []role.Permission{role.PermUserRead},
			want:        role.User,
			wantErr:     false,
		},
		{
			name:        "invalid empty role name",
			id:          uuid.New(),
			roleName:    "",
			permissions: []role.Permission{role.PermUserRead},
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := role.New(tt.id, tt.roleName, tt.permissions)

			if (err != nil) != tt.wantErr {
				t.Fatalf("role.New() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got := r.Name()
			if got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
