package models

import "errors"

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrUserForRoleNotFound    = errors.New("user not found")
	ErrRoleAlreadyAssigned    = errors.New("role already assigned to user")
	ErrRoleNotAssigned        = errors.New("role not assigned to user")
	ErrRoleInUse              = errors.New("role is in use")
	ErrSystemRoleModification = errors.New("system role cannot be modified")
	ErrRoleNameTooShort       = errors.New("role name is too short")
)

type Role struct {
	ID   int
	Name string
}

type UserRole struct {
	UserID int
	RoleID int
}

func NewRole(id int, name string) (*Role, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	return &Role{
		ID:   id,
		Name: name,
	}, nil
}

func validateName(name string) error {
	if len(name) < 3 {
		return ErrRoleNameTooShort
	}

	return nil
}
