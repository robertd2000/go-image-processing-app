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
	ErrUserBanned            = errors.New("user is banned")
	ErrUserInactive          = errors.New("user is inactive")
	ErrUserDeleted           = errors.New("user is deleted")
	ErrInvalidUserStatus     = errors.New("invalid user status")
	ErrInvalidUserRole       = errors.New("invalid user role")
	ErrInvalidUserProfile    = errors.New("invalid user profile")
	ErrInvalidUserSettings   = errors.New("invalid user settings")
	ErrInvalidUserFilter     = errors.New("invalid user filter")
	ErrInvalidUserID         = errors.New("invalid user ID")
	ErrInvalidUserFirstName  = errors.New("invalid user first name")
	ErrInvalidUserLimit      = errors.New("invalid user limit")
	ErrInvalidUserOffset     = errors.New("invalid user offset")
	ErrInvalidUserSearch     = errors.New("invalid user search")
	ErrInvalidUserSortBy     = errors.New("invalid user sort by")
	ErrInvalidUserSortOrder  = errors.New("invalid user sort order")
)
