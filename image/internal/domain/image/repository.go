package image

import (
	"context"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type Repository interface {
	Save(ctx context.Context, tx port.Tx, image *Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*Image, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Image, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
