package transformationmem

import (
	"context"
	"sync"

	"github.com/google/uuid"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
)

var _ transformDomain.Repository = (*InMemoryRepository)(nil)

type InMemoryRepository struct {
	mu sync.RWMutex

	items map[uuid.UUID]*transformDomain.Transformation

	locked map[uuid.UUID]struct{}
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		items:  make(map[uuid.UUID]*transformDomain.Transformation),
		locked: make(map[uuid.UUID]struct{}),
	}
}

func (r *InMemoryRepository) Create(
	_ context.Context,
	_ port.Tx,
	t *transformDomain.Transformation,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[t.ID()] = t

	return nil
}

func (r *InMemoryRepository) Update(
	_ context.Context,
	_ port.Tx,
	t *transformDomain.Transformation,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[t.ID()]; !ok {
		return transformDomain.ErrNotFound
	}

	r.items[t.ID()] = t

	return nil
}

func (r *InMemoryRepository) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*transformDomain.Transformation, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.items[id]
	if !ok {
		return nil, transformDomain.ErrNotFound
	}

	return t, nil
}

func (r *InMemoryRepository) GetByImageAndHash(
	_ context.Context,
	imageID uuid.UUID,
	hash string,
) (*transformDomain.Transformation, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.items {
		if t.ImageID() == imageID &&
			t.Hash() == hash {
			return t, nil
		}
	}

	return nil, transformDomain.ErrNotFound
}

func (r *InMemoryRepository) AcquireNextPending(
	_ context.Context,
	_ port.Tx,
) (*transformDomain.Transformation, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	var oldest *transformDomain.Transformation

	for _, t := range r.items {

		if t.Status() != transformDomain.StatusPending {
			continue
		}

		if _, locked := r.locked[t.ID()]; locked {
			continue
		}

		if oldest == nil || t.CreatedAt().Before(oldest.CreatedAt()) {
			oldest = t
		}
	}

	if oldest == nil {
		return nil, transformDomain.ErrNotFound
	}

	r.locked[oldest.ID()] = struct{}{}

	return oldest, nil
}
