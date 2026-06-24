package outboxmem

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type outboxRepo struct {
	mu    sync.Mutex
	data  map[uuid.UUID]*port.OutboxEvent
	clock func() time.Time
}

func NewInMemoryOutboxRepo() *outboxRepo {
	return &outboxRepo{
		data:  make(map[uuid.UUID]*port.OutboxEvent),
		clock: time.Now,
	}
}

func (r *outboxRepo) Save(ctx context.Context, tx txtx.Tx, event *port.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *event
	r.data[clone.ID] = &clone
	return nil
}

func (r *outboxRepo) FetchPending(ctx context.Context, limit int) ([]*port.OutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*port.OutboxEvent
	for _, e := range r.data {
		if e.Status == port.OutboxStatusPending {
			clone := *e
			result = append(result, &clone)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *outboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.data[id]
	if !ok {
		return nil
	}
	e.Status = port.OutboxStatusPublished
	return nil
}

func (r *outboxRepo) MarkFailed(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.data[id]
	if !ok {
		return nil
	}
	e.Status = port.OutboxStatusFailed
	return nil
}
