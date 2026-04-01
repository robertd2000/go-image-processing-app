package user

import (
	"errors"
	"net/mail"
	"strings"
)

type Username string

func NewUsername(v string) (Username, error) {
	v = strings.TrimSpace(v)

	if len(v) < 3 || len(v) > 30 {
		return "", errors.New("username must be 3-30 characters")
	}

	return Username(v), nil
}

type Email string

func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(v)

	if v == "" {
		return "", errors.New("email required")
	}

	_, err := mail.ParseAddress(v)
	if err != nil {
		return "", errors.New("invalid email")
	}

	return Email(v), nil
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
