package validation

import (
	"net/mail"

	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
)

func ValidateEmail(email string) error {
	if email == "" {
		return userDomain.ErrInvalidEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return userDomain.ErrInvalidEmail
	}

	return nil
}
