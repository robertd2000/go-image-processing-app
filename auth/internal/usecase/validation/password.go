package validation

import (
	"unicode"

	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
)

const minPasswordLength = 8

func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return userDomain.ErrInvalidPassword
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return userDomain.ErrInvalidPassword
	}

	return nil
}
