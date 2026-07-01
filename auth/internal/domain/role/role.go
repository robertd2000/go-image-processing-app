package role

import (
	"slices"

	"github.com/google/uuid"
)

type Name string

func (n Name) String() string { return string(n) }

func (n Name) IsValid() bool {
	return n == Admin || n == User
}

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

func adminPermissions() []Permission {
	return []Permission{
		PermUserRead,
		PermUserWrite,
		PermImageRead,
		PermImageWrite,
	}
}

func userPermissions() []Permission {
	return []Permission{
		PermImageRead,
		PermImageWrite,
	}
}

func New(id uuid.UUID, name Name, permissions []Permission) (*Role, error) {
	if !name.IsValid() {
		return nil, ErrInvalidRoleName
	}

	return &Role{
		id:          id,
		name:        name,
		permissions: uniquePermissions(permissions),
	}, nil
}

func FromName(id uuid.UUID, name Name) (*Role, error) {
	switch name {
	case Admin:
		return New(id, Admin, adminPermissions())
	case User:
		return New(id, User, userPermissions())
	default:
		return nil, ErrInvalidRoleName
	}
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
	result := make([]Permission, 0, len(perms))

	for _, p := range perms {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		result = append(result, p)
	}

	return result
}
