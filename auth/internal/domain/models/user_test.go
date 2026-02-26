package models_test

import (
	"testing"

	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func createUser(id int, email, password string) (*models.User, error) {
	username := "User 1"
	firstname := "User"
	lastname := "LastName 1"

	return models.NewUser(id, username, firstname, lastname, email, password)
}

func TestUserNewSuccess(t *testing.T) {
	email := "test@mail.com"
	password := "!Qwery12345678"
	id := 1

	user, err := createUser(id, email, password)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, &email)
	assert.Equal(t, user.Password, password)
	assert.Equal(t, user.ID, id)
}

func TestUserNewError(t *testing.T) {
	email := "test"
	password := "12345678"
	id := 1

	user, err := createUser(id, email, password)
	assert.Error(t, err)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, models.ErrUserValidation)
}

func TestUserNewPasswordNoUpperCaseError(t *testing.T) {
	email := "test@mail.com"
	password := "12345678"
	id := 1

	user, err := createUser(id, email, password)
	assert.Error(t, err)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, models.ErrPasswordNoUpperCase)
}

func TestUserNewPasswordTooShortError(t *testing.T) {
	email := "test@mail.com"
	password := "1234"
	id := 1

	user, err := createUser(id, email, password)
	assert.Error(t, err)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, models.ErrPasswordTooShort)
}

func TestUserNewPasswordNoDigitError(t *testing.T) {
	email := "test@mail.com"
	password := "!QweryPasswordStrong"
	id := 1

	user, err := createUser(id, email, password)
	assert.Error(t, err)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, models.ErrPasswordNoDigit)
}

func TestUserNewPasswordNoSpecialCharError(t *testing.T) {
	email := "test@mail.com"
	password := "QweryPasswordStrong1223"
	id := 1

	user, err := createUser(id, email, password)
	assert.Error(t, err)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, models.ErrPasswordNoSpecialChar)
}
