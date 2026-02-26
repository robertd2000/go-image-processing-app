package user

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrUserNotFound        = errors.New("user not found")
	ErrRoleAlreadyAssigned = errors.New("role already assigned")
	ErrRoleNotAssigned     = errors.New("role not assigned")
)
