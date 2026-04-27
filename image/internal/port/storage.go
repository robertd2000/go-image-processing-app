package port

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type Storage interface {
	Put(ctx context.Context, key uuid.UUID, reader io.Reader) error
}
