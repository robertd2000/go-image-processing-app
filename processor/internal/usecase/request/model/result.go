package model

import (
	"time"

	"github.com/google/uuid"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
)

type Result struct {
	ID        uuid.UUID
	ImageID   uuid.UUID
	Status    transformDomain.Status
	CreatedAt time.Time
}

func ToResult(t *transformDomain.Transformation) *Result {
	if t == nil {
		return nil
	}

	return &Result{
		ID:        t.ID(),
		ImageID:   t.ImageID(),
		Status:    t.Status(),
		CreatedAt: t.CreatedAt(),
	}
}
