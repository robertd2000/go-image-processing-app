package storagemem

import (
	"context"
	"io"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type spyStorage struct {
	port.Storage

	PutCalled    bool
	DeleteCalled bool
	lastKey      string
}

func (s *spyStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	s.PutCalled = true
	s.lastKey = key
	return nil
}

func (s *spyStorage) Delete(ctx context.Context, key string) error {
	s.DeleteCalled = true
	return nil
}

func NewSpyStorage() *spyStorage {
	return &spyStorage{}
}
