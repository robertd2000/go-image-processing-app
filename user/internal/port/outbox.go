package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxRepository interface {
	Create(ctx context.Context, tx Tx, e OutboxEvent) error
	GetUnprocessed(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
}

type OutboxEvent struct {
	ID          uuid.UUID
	Type        string
	Topic       string
	Key         string
	Payload     []byte
	CreatedAt   time.Time
	ProcessedAt *time.Time
}
