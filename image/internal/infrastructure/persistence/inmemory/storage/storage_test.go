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

func TestInMemoryStorage_Concurrent_ReadWrite(t *testing.T) {
	s := storagemem.NewInMemoryStorage()

	key := "key"
	initial := []byte("initial")

	_ = s.Put(context.Background(), key, bytes.NewReader(initial), int64(len(initial)), "text/plain")

	var wg sync.WaitGroup

	// writers
	for i := range 20 {
		i := i

		wg.Go(func() {

			data := fmt.Appendf(nil, "data-%d", i)
			_ = s.Put(context.Background(), key, bytes.NewReader(data), int64(len(data)), "text/plain")
		})
	}

	// readers
	for range 50 {

		wg.Go(func() {

			r, err := s.Get(context.Background(), key)
			if err == nil {
				_, _ = io.ReadAll(r)
			}
		})
	}

	wg.Wait()
}

func TestStorage_GetURL(t *testing.T) {
	ctx := context.Background()

	s := storagemem.NewInMemoryStorage()

	key := "test/image.jpg"
	data := []byte("fake-image-data")

	// --- prepare ---
	err := s.Put(ctx, key, bytes.NewReader(data), int64(len(data)), "image/jpeg")
	assert.NoError(t, err)

	// --- act ---
	url, err := s.GetURL(ctx, key)

	// --- assert ---
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, key)
}

func TestStorage_GetURL_NotFound(t *testing.T) {
	ctx := context.Background()

	s := storagemem.NewInMemoryStorage()

	// --- act ---
	url, err := s.GetURL(ctx, "not-exists")

	// --- assert ---
	assert.Error(t, err)
	assert.Empty(t, url)
}
