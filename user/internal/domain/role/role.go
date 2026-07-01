package role

import (
	"errors"
	"slices"
)

type Name string

var ErrInvalidRoleName = errors.New("invalid role name")

const (
	Admin Name = "ADMIN"
	User  Name = "USER"
)

func (n Name) String() string {
	return string(n)
}

func (n Name) IsValid() bool {
	return n == Admin || n == User
}

type Permission string

const (
	PermUserRead   Permission = "user:read"
	PermUserWrite  Permission = "user:write"
	PermImageRead  Permission = "image:read"
	PermImageWrite Permission = "image:write"
)

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

type Role struct {
	name        Name
	permissions []Permission
}

func New(name Name, permissions []Permission) (*Role, error) {
	if !name.IsValid() {
		return nil, ErrInvalidRoleName
	}

	return &Role{
		name:        name,
		permissions: uniquePermissions(permissions),
	}, nil
}

func FromName(name Name) (*Role, error) {
	switch name {
	case Admin:
		return New(Admin, adminPermissions())
	case User:
		return New(User, userPermissions())
	default:
		return nil, ErrInvalidRoleName
	}
}

func (r Role) Name() Name {
	return r.name
}

func (r Role) HasPermission(p Permission) bool {
	return slices.Contains(r.permissions, p)
}

func (r Role) Permissions() []Permission {
	cp := make([]Permission, len(r.permissions))
	copy(cp, r.permissions)
	return cp
}
func uniquePermissions(perms []Permission) []Permission {
	seen := make(map[Permission]struct{}, len(perms))
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
