package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	id := uuid.New()

	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	if user.ID() != id {
		t.Errorf("expected id %v, got %v", id, user.ID())
	}

	if user.Username() != username {
		t.Errorf("username mismatch")
	}

	if user.Email() != email {
		t.Errorf("email mismatch")
	}

	if user.Profile() == nil {
		t.Errorf("profile should be initialized")
	}

	if user.Settings() == nil {
		t.Errorf("settings should be initialized")
	}
}

func TestUser_ChangeUsername(t *testing.T) {
	id := uuid.New()
	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	newUsername, _ := NewUsername("newname")

	err := user.ChangeUsername(newUsername)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Username() != newUsername {
		t.Errorf("username not updated")
	}
}

func TestUser_ChangeUsername_Banned(t *testing.T) {
	id := uuid.New()
	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	user.status = StatusBanned

	newUsername, _ := NewUsername("newname")

	err := user.ChangeUsername(newUsername)
	if err == nil {
		t.Fatalf("expected error for banned user")
	}
}

func TestUser_UpdateName(t *testing.T) {
	id := uuid.New()
	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	user.UpdateName("John", "Doe")

	// TODO: Add getters for first and last name to verify the update
}

func TestUser_Deactivate(t *testing.T) {
	id := uuid.New()
	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	user.Deactivate()

	if user.status != StatusInactive {
		t.Errorf("expected inactive status")
	}

	if user.deletedAt == nil {
		t.Errorf("deletedAt should be set")
	}
}

func TestUser_UpdateAvatar(t *testing.T) {
	id := uuid.New()
	username, _ := NewUsername("john")
	email, _ := NewEmail("john@mail.com")

	user := NewUser(id, username, email)

	url := "http://avatar.com/img.png"

	user.UpdateAvatar(&url)

	if user.avatarURL == nil || *user.avatarURL != url {
		t.Errorf("avatar not updated")
	}
}
