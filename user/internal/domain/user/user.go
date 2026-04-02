package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id uuid.UUID

	username Username
	email    Email

	firstName string
	lastName  string

	avatarURL *string

	status UserStatus
	role   UserRole

	profile  *UserProfile
	settings *UserSettings

	lastSeenAt *time.Time

	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewUser(
	id uuid.UUID,
	username Username,
	email Email,
) *User {
	now := time.Now()

	return &User{
		id:        id,
		username:  username,
		email:     email,
		status:    StatusActive,
		role:      RoleUser,
		profile:   NewProfile(),
		settings:  NewSettings(),
		createdAt: now,
		updatedAt: now,
	}
}

func (u *User) ChangeUsername(username Username) error {
	if u.status == StatusBanned {
		return errors.New("banned user cannot change username")
	}

	u.username = username
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateName(first, last string) {
	u.firstName = first
	u.lastName = last
	u.updatedAt = time.Now()
}

func (u *User) ChangeFirstName(first string) error {
	if u.status == StatusBanned {
		return errors.New("banned user cannot change first name")
	}

	u.firstName = first
	u.updatedAt = time.Now()
	return nil
}

func (u *User) ChangeLastname(last string) error {
	if u.status == StatusBanned {
		return errors.New("banned user cannot change last name")
	}
	u.lastName = last
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateAvatar(url *string) {
	u.avatarURL = url
	u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
	now := time.Now()
	u.status = StatusInactive
	u.deletedAt = &now
	u.updatedAt = now
}

func (u *User) UpdateLastSeen() {
	now := time.Now()
	u.lastSeenAt = &now
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) ChangeEmail(email Email) error {
	if u.status == StatusBanned {
		return errors.New("banned user cannot change email")
	}

	u.email = email
	u.updatedAt = time.Now()
	return nil
}

func (u *User) Role() UserRole {
	return u.role
}

func (u *User) Status() UserStatus {
	return u.status
}

func (u *User) Profile() *UserProfile {
	return u.profile
}

func (u *User) Settings() *UserSettings {
	return u.settings
}

func (u *User) AvatarURL() *string {
	return u.avatarURL
}

func (u *User) FirstName() string {
	return u.firstName
}

func (u *User) LastName() string {
	return u.lastName
}

func NewUserFromDB(
	id uuid.UUID,
	username string,
	email string,
	status string,
	role string,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:        id,
		username:  Username(username),
		email:     Email(email),
		status:    UserStatus(status),
		role:      UserRole(role),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
