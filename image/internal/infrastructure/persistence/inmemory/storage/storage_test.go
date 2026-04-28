package storagemem_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	storagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/storage"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_PutGet(t *testing.T) {
	s := storagemem.NewInMemoryStorage()

	key := "test-key"
	data := []byte("hello")

	err := s.Put(context.Background(), key, bytes.NewReader(data), int64(len(data)), "text/plain")
	assert.NoError(t, err)

	r, err := s.Get(context.Background(), key)
	assert.NoError(t, err)

	got, err := io.ReadAll(r)
	assert.NoError(t, err)

	assert.Equal(t, data, got)
}

func TestInMemoryStorage_Delete(t *testing.T) {
	s := storagemem.NewInMemoryStorage()

	key := "key"

	_ = s.Put(context.Background(), key, bytes.NewReader([]byte("data")), 4, "text/plain")

	err := s.Delete(context.Background(), key)
	assert.NoError(t, err)

	_, err = s.Get(context.Background(), key)
	assert.Error(t, err)
}

func TestInMemoryStorage_Concurrent_PutGet(t *testing.T) {
	s := storagemem.NewInMemoryStorage()

	const workers = 100

	var wg sync.WaitGroup

	for i := range workers {
		i := i

		wg.Go(func() {

			key := fmt.Sprintf("key-%d", i)
			data := fmt.Appendf(nil, "data-%d", i)

			err := s.Put(context.Background(), key, bytes.NewReader(data), int64(len(data)), "text/plain")
			assert.NoError(t, err)

			r, err := s.Get(context.Background(), key)
			assert.NoError(t, err)

			got, err := io.ReadAll(r)
			assert.NoError(t, err)

			assert.Equal(t, data, got)
		})
	}

	wg.Wait()
}

func TestInMemoryStorage_Concurrent_SameKey(t *testing.T) {
	s := storagemem.NewInMemoryStorage()

	const workers = 50

	var wg sync.WaitGroup

	for i := range workers {
		i := i

		wg.Go(func() {

			data := fmt.Appendf(nil, "data-%d", i)

			err := s.Put(context.Background(), "shared-key", bytes.NewReader(data), int64(len(data)), "text/plain")
			assert.NoError(t, err)
		})
	}

	wg.Wait()

	r, err := s.Get(context.Background(), "shared-key")
	assert.NoError(t, err)

	got, err := io.ReadAll(r)
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(string(got), "data-"))
}
