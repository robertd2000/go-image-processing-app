package user

import "errors"

var (
	ErrInvalidUsername       = errors.New("invalid username: username must be 3-30 characters")
	ErrInvalidEmail          = errors.New("invalid email: email must be a valid email address")
	ErrEmailRequired         = errors.New("email required")
	ErrUsernameRequired      = errors.New("username required")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
)
