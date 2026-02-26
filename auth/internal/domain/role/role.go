// Package role
package role

import (
	"slices"

	"github.com/google/uuid"
)

type Name string

const (
	Admin Name = "ADMIN"
	User  Name = "USER"
)

type Permission string

const (
	PermUserRead   Permission = "user:read"
	PermUserWrite  Permission = "user:write"
	PermImageRead  Permission = "image:read"
	PermImageWrite Permission = "image:write"
)

type Role struct {
	id          uuid.UUID
	name        Name
	permissions []Permission
}

func New(id uuid.UUID, name Name, permissions []Permission) (*Role, error) {
	if name == "" {
		return nil, ErrInvalidRoleName
	}

	return &Role{
		id:          id,
		name:        name,
		permissions: uniquePermissions(permissions),
	}, nil
}

func (r *Role) ID() uuid.UUID {
	return r.id
}

func (r *Role) Name() Name {
	return r.name
}

func (r *Role) HasPermission(p Permission) bool {
	return slices.Contains(r.permissions, p)
}

func (r *Role) Permissions() []Permission {
	cp := make([]Permission, len(r.permissions))
	copy(cp, r.permissions)
	return cp
}

func uniquePermissions(perms []Permission) []Permission {
	seen := make(map[Permission]struct{})
	var result []Permission

	for _, p := range perms {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			result = append(result, p)
		}
	}

	return result
}
