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

func FromName(name string) (Role, error) {
	switch Name(name) {
	case Admin:
		return Role{
			name:        Admin,
			permissions: adminPermissions(),
		}, nil
	case User:
		return Role{
			name:        User,
			permissions: userPermissions(),
		}, nil
	default:
		return Role{}, ErrInvalidRoleName
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
