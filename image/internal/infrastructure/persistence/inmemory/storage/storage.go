package storagemem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

type object struct {
	data        []byte
	contentType string
}

type storage struct {
	mu   sync.RWMutex
	data map[string]object
}

func NewStorage() *storage {
	return &storage{
		data: make(map[string]object),
	}
}

func (s *storage) Put(
	ctx context.Context,
	key string,
	r io.Reader,
	size int64,
	contentType string,
) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = object{
		data:        data,
		contentType: contentType,
	}
	return nil
}

func (s *storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("object not found")
	}

	return io.NopCloser(bytes.NewReader(data.data)), nil
}

func (s *storage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}
