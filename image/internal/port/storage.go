package port

import (
	"context"
	"io"
)

type Storage interface {
	Put(
		ctx context.Context,
		key string,
		r io.Reader,
		size int64,
		contentType string,
	) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	GetURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}
