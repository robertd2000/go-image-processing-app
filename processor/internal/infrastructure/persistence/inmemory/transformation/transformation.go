package transformationmem

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/domain/tx"
)

type repo struct {
	mu    sync.Mutex
	data  map[uuid.UUID]*transformation.Transformation
	clock func() time.Time
}

func NewInMemoryTransformRepo() *repo {
	return &repo{
		data:  make(map[uuid.UUID]*transformation.Transformation),
		clock: time.Now,
	}
}

func (r *repo) Create(ctx context.Context, tx txtx.Tx, t *transformation.Transformation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *t
	r.data[clone.ID()] = &clone
	return nil
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*transformation.Transformation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.data[id]
	if !ok {
		return nil, transformation.ErrNotFound
	}
	clone := *t
	return &clone, nil
}

func (r *repo) GetByImageAndHash(ctx context.Context, imageID uuid.UUID, hash string) (*transformation.Transformation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.data {
		if t.ImageID() == imageID && t.Hash() == hash {
			clone := *t
			return &clone, nil
		}
	}
	return nil, transformation.ErrNotFound
}

func ptr[T any](v T) *T { return &v }

func (r *repo) Seed(ctx context.Context, t *transformation.Transformation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t2, err := transformation.RestoreTransformation(
		t.ID(), t.ImageID(),
		json.RawMessage(`{}`), t.Hash(),
		t.Status(), "", "",
		ptr(time.Now()), nil,
		0, r.clock(),
	)
	if err != nil {
		panic(err)
	}
	r.data[t2.ID()] = t2
}
