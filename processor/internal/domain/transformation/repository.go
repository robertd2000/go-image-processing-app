package transformation

import (
	"context"

	"github.com/google/uuid"
	tx "github.com/robertd2000/go-image-processing-app/processor/internal/port"
)

type Repository interface {
	Create(ctx context.Context, tx tx.Tx, t *Transformation) error
	Update(ctx context.Context, tx tx.Tx, t *Transformation) error

	GetByID(ctx context.Context, id uuid.UUID) (*Transformation, error)
	GetByImageAndHash(
		ctx context.Context,
		imageID uuid.UUID,
		hash string,
	) (*Transformation, error)
	GetPending(ctx context.Context, limit int) ([]*Transformation, error)
}
