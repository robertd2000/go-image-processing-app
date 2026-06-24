package outboxpg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"go.uber.org/zap"
)

type outboxRepository struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewOutboxRepository(pool *pgxpool.Pool, logger *zap.SugaredLogger) *outboxRepository {
	return &outboxRepository{pool: pool, logger: logger}
}

func (r *outboxRepository) Save(ctx context.Context, tx txtx.Tx, event *port.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("outbox save: marshal payload: %w", err)
	}
	return tx.Exec(ctx, query, event.ID, event.AggregateID, event.EventType, payload, string(event.Status), event.CreatedAt)
}

func (r *outboxRepository) FetchPending(ctx context.Context, limit int) ([]*port.OutboxEvent, error) {
	query := `
		SELECT id, aggregate_id, event_type, payload, status, created_at
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox fetch pending: %w", err)
	}
	defer rows.Close()

	var result []*port.OutboxEvent
	for rows.Next() {
		var (
			id          uuid.UUID
			aggregateID uuid.UUID
			eventType   string
			payload     []byte
			status      string
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &aggregateID, &eventType, &payload, &status, &createdAt); err != nil {
			return nil, fmt.Errorf("outbox scan: %w", err)
		}
		result = append(result, &port.OutboxEvent{
			ID:          id,
			AggregateID: aggregateID,
			EventType:   eventType,
			Payload:     payload,
			Status:      port.OutboxEventStatus(status),
			CreatedAt:   createdAt,
		})
	}
	return result, nil
}

func (r *outboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET status = 'published' WHERE id = $1`, id)
	if err != nil {
		r.logger.Errorw("outbox mark published failed", "id", id, "error", err)
	}
	return err
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET status = 'failed' WHERE id = $1`, id)
	if err != nil {
		r.logger.Errorw("outbox mark failed", "id", id, "error", err)
	}
	return err
}
