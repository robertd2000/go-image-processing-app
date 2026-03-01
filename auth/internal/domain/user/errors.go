package user

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrRoleAlreadyAssigned = errors.New("role already assigned")
	ErrRoleNotAssigned     = errors.New("role not assigned")
	ErrInvalidUsername     = errors.New("invalid username")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
)
