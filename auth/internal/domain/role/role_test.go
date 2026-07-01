package role_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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
		},
		{
			name:        "user role",
			id:          uuid.New(),
			roleName:    role.User,
			permissions: []role.Permission{role.PermUserRead},
			want:        role.User,
		},
		{
			name:        "invalid empty role name",
			id:          uuid.New(),
			roleName:    "",
			permissions: []role.Permission{role.PermUserRead},
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

			assert.Equal(t, tt.want, r.Name())
		})
	}
}

func TestFromName(t *testing.T) {
	tests := []struct {
		name      string
		roleName  role.Name
		wantPerms []role.Permission
		wantErr   bool
	}{
		{
			name:     "admin",
			roleName: role.Admin,
			wantPerms: []role.Permission{
				role.PermUserRead,
				role.PermUserWrite,
				role.PermImageRead,
				role.PermImageWrite,
			},
		},
		{
			name:     "user",
			roleName: role.User,
			wantPerms: []role.Permission{
				role.PermImageRead,
				role.PermImageWrite,
			},
		},
		{
			name:     "invalid role",
			roleName: role.Name("UNKNOWN"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()

			r, err := role.FromName(id, tt.roleName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("role.FromName() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			assert.Equal(t, id, r.ID())
			assert.Equal(t, tt.roleName, r.Name())
			assert.Equal(t, tt.wantPerms, r.Permissions())
		})
	}
}

func TestRole_HasPermission(t *testing.T) {
	r, err := role.FromName(uuid.New(), role.Admin)
	assert.NoError(t, err)

	assert.True(t, r.HasPermission(role.PermUserRead))
	assert.True(t, r.HasPermission(role.PermUserWrite))
	assert.True(t, r.HasPermission(role.PermImageRead))
	assert.True(t, r.HasPermission(role.PermImageWrite))

	assert.False(t, r.HasPermission(role.Permission("unknown")))
}

func TestRole_PermissionsReturnsCopy(t *testing.T) {
	r, err := role.FromName(uuid.New(), role.Admin)
	assert.NoError(t, err)

	perms := r.Permissions()
	perms[0] = role.Permission("modified")

	assert.NotEqual(t, perms, r.Permissions())
}
