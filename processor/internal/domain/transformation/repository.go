package transformation

import (
	"context"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/domain/tx"
)

type Repository interface {
	Create(ctx context.Context, tx txtx.Tx, t *Transformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transformation, error)
	GetByImageAndHash(ctx context.Context, imageID uuid.UUID, hash string) (*Transformation, error)
}
