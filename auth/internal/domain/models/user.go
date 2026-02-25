// Package models contains user models
package models

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidUser           = errors.New("invalid user")
	ErrUserValidation        = errors.New("validation error")
	ErrUserAlreadyExist      = errors.New("user already exist")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPassword       = errors.New("password must be at least 6 characters")
	ErrUserExists            = errors.New("user already exists")
	ErrUserNotExists         = errors.New("user not exists")
	ErrNotFound              = errors.New("user not found")
	ErrPasswordHash          = errors.New("failed to hash password")
	ErrPasswordWrong         = errors.New("wrong password")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordNoUpperCase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecialChar = errors.New("password must contain at least one special character")
)

type User struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	Email     *string
	Password  string
	Enabled   bool

	Roles []Role

	CreatedAt  time.Time
	ModifiedAt *time.Time
	DeletedAt  *time.Time
}

func NewUser(id int, username, firstname, lastname, email, password string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Username:  username,
		FirstName: firstname,
		LastName:  lastname,
		Email:     &email,
		Password:  password,
		Enabled:   true,
		Roles:     []Role{},
		CreatedAt: time.Now(),
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrUserValidation)
	}

	re := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if match, _ := regexp.MatchString(re, email); !match {
		return fmt.Errorf("%w: invalid email format", ErrUserValidation)
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	match, _ := regexp.MatchString(`[A-Z]`, password)
	if !match {
		return ErrPasswordNoUpperCase
	}

	match, _ = regexp.MatchString(`[0-9]`, password)
	if !match {
		return ErrPasswordNoDigit
	}

	match, _ = regexp.MatchString(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};:'"\\|,.<>\/?]`, password)
	if !match {
		return ErrPasswordNoSpecialChar
	}

	return nil
}
