package role

import (
	"context"

	"github.com/google/uuid"

	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

type Repository interface {
	ByID(ctx context.Context, tx txtx.Tx, id uuid.UUID) (*Role, error)
	ByName(ctx context.Context, tx txtx.Tx, name Name) (*Role, error)
}
