package transformationmem

import (
	"context"
	"sync"

	"github.com/google/uuid"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	txmem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/txmanager"
	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/port"
)

var _ transformDomain.Repository = (*Repository)(nil)

type Repository struct {
	mu sync.RWMutex

	byID   map[uuid.UUID]*transformDomain.Transformation
	byHash map[string]uuid.UUID
}

func NewRepository() *Repository {
	return &Repository{
		byID:   make(map[uuid.UUID]*transformDomain.Transformation),
		byHash: make(map[string]uuid.UUID),
	}
}

func hashKey(imageID uuid.UUID, hash string) string {
	return imageID.String() + ":" + hash
}

func (r *Repository) Create(
	ctx context.Context,
	tx txtx.Tx,
	t *transformDomain.Transformation,
) error {

	if fakeTx, ok := tx.(*txmem.FakeTx); ok {
		fakeTx.OnCommit(func(ctx context.Context) error {
			r.mu.Lock()
			defer r.mu.Unlock()

			r.byID[t.ID()] = t
			r.byHash[hashKey(t.ImageID(), t.Hash())] = t.ID()

			return nil
		})

		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[t.ID()] = t
	r.byHash[hashKey(t.ImageID(), t.Hash())] = t.ID()

	return nil
}

func (r *Repository) Update(
	ctx context.Context,
	tx txtx.Tx,
	t *transformDomain.Transformation,
) error {

	if fakeTx, ok := tx.(*txmem.FakeTx); ok {
		fakeTx.OnCommit(func(ctx context.Context) error {
			r.mu.Lock()
			defer r.mu.Unlock()

			if _, ok := r.byID[t.ID()]; !ok {
				return transformDomain.ErrNotFound
			}

			r.byID[t.ID()] = t

			return nil
		})

		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[t.ID()]; !ok {
		return transformDomain.ErrNotFound
	}

	r.byID[t.ID()] = t

	return nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*transformDomain.Transformation, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byID[id]
	if !ok {
		return nil, transformDomain.ErrNotFound
	}

	return t, nil
}

func (r *Repository) GetByImageAndHash(
	ctx context.Context,
	imageID uuid.UUID,
	hash string,
) (*transformDomain.Transformation, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byHash[hashKey(imageID, hash)]
	if !ok {
		return nil, transformDomain.ErrNotFound
	}

	return r.byID[id], nil
}

func (r *Repository) GetPending(
	ctx context.Context,
	limit int,
) ([]*transformDomain.Transformation, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*transformDomain.Transformation, 0, limit)

	for _, t := range r.byID {
		if t.Status() != transformDomain.StatusPending {
			continue
		}

		result = append(result, t)

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}
