// Package user
package user

import "github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"

type User struct {
	id           int64
	email        string
	passwordHash string
	roles        []*role.Role
}

func New(id int64, email, passwordHash string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}

	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		roles:        []*role.Role{},
	}, nil
}

func (u *User) ID() int64 {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Roles() []*role.Role {
	return u.roles
}

func (u *User) AddRole(r *role.Role) error {
	for _, existing := range u.roles {
		if existing.ID() == r.ID() {
			return ErrRoleAlreadyAssigned
		}
	}
	u.roles = append(u.roles, r)
	return nil
}

func (u *User) RemoveRole(roleID int64) error {
	for i, r := range u.roles {
		if r.ID() == roleID {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			return nil
		}
	}
	return ErrRoleNotAssigned
}

func (u *User) HasPermission(p role.Permission) bool {
	for _, r := range u.roles {
		if r.HasPermission(p) {
			return true
		}
	}
	return false
}
