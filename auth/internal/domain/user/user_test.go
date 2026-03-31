package user_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

var (
	username     = "username"
	firtname     = "First"
	lastname     = "Last"
	email        = "user@example.com"
	passwordHash = "!!!!123PasswordHashSecure11111!?"
)

func createTestUser(t *testing.T) *user.AuthUser {
	t.Helper()
	userID := uuid.New()
	u, err := user.NewAuthUser(userID, username, &email, passwordHash)
	assert.NoError(t, err)
	return u
}

func TestAddRole_Success(t *testing.T) {
	u := createTestUser(t)
	r, _ := role.New(uuid.New(), "admin", []role.Permission{"read", "write"})

	err := u.AddRole(*r)
	assert.NoError(t, err)
	assert.Len(t, u.Roles(), 1)
	assert.Equal(t, "admin", string(u.Roles()[0].Name()))
}

func TestAddRole_AlreadyAssigned(t *testing.T) {
	u := createTestUser(t)
	r, _ := role.New(uuid.New(), "admin", nil)

	_ = u.AddRole(*r)
	err := u.AddRole(*r)
	assert.ErrorIs(t, err, user.ErrRoleAlreadyAssigned)
}

func TestRemoveRole_Success(t *testing.T) {
	u := createTestUser(t)
	r, _ := role.New(uuid.New(), "admin", nil)

	_ = u.AddRole(*r)
	err := u.RemoveRole(r.ID())
	assert.NoError(t, err)
	assert.Len(t, u.Roles(), 0)
}

func TestRemoveRole_NotAssigned(t *testing.T) {
	u := createTestUser(t)
	r, _ := role.New(uuid.New(), "admin", nil)

	err := u.RemoveRole(r.ID())
	assert.ErrorIs(t, err, user.ErrRoleNotAssigned)
}
