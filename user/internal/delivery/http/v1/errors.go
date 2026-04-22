package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
)

type httpError struct {
	Status  int
	Message string
}

func mapError(err error) httpError {
	switch {
	case errors.Is(err, userDomain.ErrUserNotFound):
		return httpError{http.StatusNotFound, "user not found"}

	case errors.Is(err, userDomain.ErrUserAlreadyExists),
		errors.Is(err, userDomain.ErrEmailAlreadyExists),
		errors.Is(err, userDomain.ErrUsernameAlreadyExists):
		return httpError{http.StatusConflict, err.Error()}

	case errors.Is(err, userDomain.ErrInvalidEmail),
		errors.Is(err, userDomain.ErrInvalidUsername),
		errors.Is(err, userDomain.ErrInvalidUserID),
		errors.Is(err, userDomain.ErrInvalidUserProfile),
		errors.Is(err, userDomain.ErrInvalidUserSettings):
		return httpError{http.StatusBadRequest, err.Error()}

	case errors.Is(err, userDomain.ErrUserBanned),
		errors.Is(err, userDomain.ErrUserInactive),
		errors.Is(err, userDomain.ErrUserDeleted):
		return httpError{http.StatusForbidden, err.Error()}

	default:
		return httpError{http.StatusInternalServerError, "internal server error"}
	}
}

func respondError(c *gin.Context, err error) {
	e := mapError(err)
	c.JSON(e.Status, gin.H{"error": e.Message})
}
