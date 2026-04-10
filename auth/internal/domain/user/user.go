// Package user
package user

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	roleDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
)

type AuthUser struct {
	id           uuid.UUID
	username     string
	email        *string
	passwordHash string
	status       string
	roles        []roleDomain.Role

	createdAt time.Time
}

func NewAuthUser(
	userID uuid.UUID,
	username string,
	email *string,
	passwordHash string,
) (*AuthUser, error) {
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

	return &AuthUser{
		id:           userID,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		roles:        []roleDomain.Role{},
		createdAt:    now,
		status:       "active",
	}, nil
}

func (u *AuthUser) ID() uuid.UUID { return u.id }

func (u *AuthUser) Username() string     { return u.username }
func (u *AuthUser) Email() *string       { return u.email }
func (u *AuthUser) PasswordHash() string { return u.passwordHash }
func (u *AuthUser) Status() string       { return u.status }
func (u *AuthUser) Roles() []roleDomain.Role {
	rolesCopy := make([]roleDomain.Role, len(u.roles))
	copy(rolesCopy, u.roles)
	return rolesCopy
}
func (u *AuthUser) CreatedAt() time.Time { return u.createdAt }

func (u *AuthUser) AddRole(r roleDomain.Role) error {
	for _, existing := range u.roles {
		if existing.ID() == r.ID() {
			return ErrRoleAlreadyAssigned
		}
	}
	u.roles = append(u.roles, r)
	return nil
}

func (u *AuthUser) RemoveRole(roleID uuid.UUID) error {
	for i, r := range u.roles {
		if r.ID() == roleID {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			return nil
		}
	}
	return ErrRoleNotAssigned
}

func (u *AuthUser) HasPermission(p roleDomain.Permission) bool {
	for _, r := range u.roles {
		if r.HasPermission(p) {
			return true
		}
	}
	return false
}

func (u *AuthUser) UpdateEmail(e string) error {
	if err := validateEmail(&e); err != nil {
		return err
	}

	u.email = &e
	return nil
}

func (u *AuthUser) UpdateStatus(s string) error {
	s = strings.TrimSpace(s)

	u.status = s
	return nil
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

func NewUserFromDB(
	id uuid.UUID,
	username string,
	email *string,
	passwordHash string,
	status string,
	createdAt time.Time,
	roles []roleDomain.Role,

) *AuthUser {
	return &AuthUser{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		status:       status,
		createdAt:    createdAt,
		roles:        roles,
	}
}
