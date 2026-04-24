package image

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, image *Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*Image, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Image, error)
}
