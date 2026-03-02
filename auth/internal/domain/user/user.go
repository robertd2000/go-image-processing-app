// Package user
package user

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

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

func NewUser(
	userID uuid.UUID,
	username, firstName, lastName string,
	email *string,
	passwordHash string,
) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePasswordHash(passwordHash); err != nil {
		return nil, err
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

func CreateUser(
	username, firstName, lastName string,
	email *string,
	passwordHash string,
) (*User, error) {
	return NewUser(uuid.New(), username, firstName, lastName, email, passwordHash)
}

func (u *User) ID() uuid.UUID { return u.id }

func (u *User) Username() string       { return u.username }
func (u *User) FirstName() string      { return u.firstName }
func (u *User) LastName() string       { return u.lastName }
func (u *User) Email() string          { return *u.email }
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

func (u *User) Clone() *User {
	if u == nil {
		return nil
	}

	clone := *u

	if u.email != nil {
		emailCopy := *u.email
		clone.email = &emailCopy
	}

	if u.modifiedAt != nil {
		t := *u.modifiedAt
		clone.modifiedAt = &t
	}

	if u.deletedAt != nil {
		t := *u.deletedAt
		clone.deletedAt = &t
	}

	if u.roles != nil {
		rolesCopy := make([]role.Role, len(u.roles))
		copy(rolesCopy, u.roles)
		clone.roles = rolesCopy
	}

	return &clone
}

func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return ErrInvalidUsername
	}

	if !utf8.ValidString(username) {
		return ErrInvalidUsername
	}

	if len(username) < 3 || len(username) > 50 {
		return ErrInvalidUsername
	}

	return nil
}

func validateEmail(email *string) error {
	if email == nil {
		return nil
	}

	e := strings.TrimSpace(*email)
	if e == "" {
		return ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(e); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

func validatePasswordHash(hash string) error {
	hash = strings.TrimSpace(hash)

	if hash == "" {
		return ErrInvalidPasswordHash
	}

	if len(hash) < 20 {
		return ErrInvalidPasswordHash
	}

	return nil
}
