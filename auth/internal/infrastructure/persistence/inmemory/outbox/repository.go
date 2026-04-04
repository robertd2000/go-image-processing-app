package outboxmem

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type Repository struct {
	mu     sync.Mutex
	events map[uuid.UUID]port.OutboxEvent
}

func NewRepository() *Repository {
	return &Repository{
		events: make(map[uuid.UUID]port.OutboxEvent),
	}
}

func (r *Repository) Create(ctx context.Context, tx port.Tx, e port.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e.CreatedAt = time.Now()
	r.events[e.ID] = e

	return nil
}

func (r *Repository) GetUnprocessed(ctx context.Context, limit int) ([]port.OutboxEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []port.OutboxEvent

	for _, e := range r.events {
		if e.ProcessedAt == nil {
			result = append(result, e)
		}
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func (r *Repository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	e := r.events[id]
	now := time.Now()
	e.ProcessedAt = &now
	r.events[id] = e

	return nil
}
