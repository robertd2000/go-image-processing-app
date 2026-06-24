package image

import (
	"context"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
)

type Repository interface {
	Save(ctx context.Context, tx txtx.Tx, image *Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*Image, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Image, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
