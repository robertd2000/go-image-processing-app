// Package user
package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
)

type User struct {
	id           uuid.UUID
	username     string
	firstName    string
	lastName     string
	email        *string
	passwordHash string
	enabled      bool
	roles        []role.Role

	createdAt  time.Time
	modifiedAt *time.Time
	deletedAt  *time.Time
}

func New(
	username, firstName, lastName string,
	email *string,
	passwordHash string,
) (*User, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}
	if passwordHash == "" {
		return nil, ErrInvalidPasswordHash
	}

	now := time.Now()

	return &User{
		id:           uuid.New(),
		username:     username,
		firstName:    firstName,
		lastName:     lastName,
		email:        email,
		passwordHash: passwordHash,
		enabled:      true,
		roles:        []role.Role{},
		createdAt:    now,
	}, nil
}

func (u *User) ID() uuid.UUID { return u.id }

func (u *User) Username() string       { return u.username }
func (u *User) FirstName() string      { return u.firstName }
func (u *User) LastName() string       { return u.lastName }
func (u *User) Email() *string         { return u.email }
func (u *User) PasswordHash() string   { return u.passwordHash }
func (u *User) Enabled() bool          { return u.enabled }
func (u *User) Roles() []role.Role     { return u.roles }
func (u *User) CreatedAt() time.Time   { return u.createdAt }
func (u *User) ModifiedAt() *time.Time { return u.modifiedAt }
func (u *User) DeletedAt() *time.Time  { return u.deletedAt }

func (u *User) AddRole(r role.Role) error {
	for _, existing := range u.roles {
		if existing.ID() == r.ID() {
			return ErrRoleAlreadyAssigned
		}
	}
	u.roles = append(u.roles, r)
	u.touch()
	return nil
}

func (u *User) RemoveRole(roleID uuid.UUID) error {
	for i, r := range u.roles {
		if r.ID() == roleID {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			u.touch()
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

func (u *User) Enable() {
	u.enabled = true
	u.touch()
}

func (u *User) Disable() {
	u.enabled = false
	u.touch()
}

func (u *User) UpdateEmail(e string) {
	u.email = &e
	u.touch()
}

func (u *User) touch() {
	now := time.Now()
	u.modifiedAt = &now
}
