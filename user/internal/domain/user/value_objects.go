package user

import (
	"net/mail"
	"strings"
)

type Username string

func NewUsername(v string) (Username, error) {
	v = strings.TrimSpace(v)

	if len(v) < 3 || len(v) > 30 {
		return "", ErrInvalidUsername
	}

	return Username(v), nil
}

func (u Username) String() string {
	return string(u)
}

type Email string

func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(v)

	if v == "" {
		return "", ErrEmailRequired
	}

	_, err := mail.ParseAddress(v)
	if err != nil {
		return "", ErrInvalidEmail
	}

	return Email(v), nil
}

func (e Email) String() string {
	return string(e)
}

type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned   UserStatus = "banned"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)
