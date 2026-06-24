package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type httpError struct {
	Status  int
	Message string
}

func mapError(err error) httpError {
	switch {
	case errors.Is(err, imageDomain.ErrNotFound):
		return httpError{http.StatusNotFound, "image not found"}

	case errors.Is(err, imageDomain.ErrAlreadyExists):
		return httpError{http.StatusConflict, err.Error()}

	case errors.Is(err, imageDomain.ErrInvalidUserID),
		errors.Is(err, imageDomain.ErrInvalidImageDimensions),
		errors.Is(err, imageDomain.ErrInvalidImageSize),
		errors.Is(err, imageDomain.ErrInvalidImageMissingMimeType),
		errors.Is(err, imageDomain.ErrInvalidImageMissingReader),
		errors.Is(err, imageDomain.ErrInvalidPagination),
		errors.Is(err, imageDomain.ErrExtractMetadata):
		return httpError{http.StatusBadRequest, err.Error()}

	default:
		return httpError{http.StatusInternalServerError, "internal server error"}
	}
}

func respondError(c *gin.Context, err error) {
	e := mapError(err)
	c.JSON(e.Status, gin.H{"error": e.Message})
}
