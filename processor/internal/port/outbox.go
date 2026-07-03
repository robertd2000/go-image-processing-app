package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/domain/tx"
)

type OutboxEventStatus string

const (
	OutboxStatusPending   OutboxEventStatus = "pending"
	OutboxStatusPublished OutboxEventStatus = "published"
	OutboxStatusFailed    OutboxEventStatus = "failed"
)

type OutboxEvent struct {
	ID          uuid.UUID
	AggregateID uuid.UUID
	EventType   string
	Payload     json.RawMessage
	Status      OutboxEventStatus
	CreatedAt   time.Time
}

type OutboxRepository interface {
	Save(ctx context.Context, tx txtx.Tx, event *OutboxEvent) error
	FetchPending(ctx context.Context, limit int) ([]*OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
}
