package role

import "errors"

var (
	ErrInvalidRoleName = errors.New("invalid role name")
	ErrRoleNotFound    = errors.New("role not found")
)
