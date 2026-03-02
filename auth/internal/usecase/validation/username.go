package validation

import (
	"regexp"

	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return userDomain.ErrInvalidUsername
	}

	return nil
}
