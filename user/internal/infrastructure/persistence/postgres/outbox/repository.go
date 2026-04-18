package outboxpg

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/user/internal/port"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, tx port.Tx, e port.OutboxEvent) error {
	err := tx.Exec(ctx, `
		INSERT INTO outbox_events (id, event_type, topic, key, payload)
		VALUES ($1, $2, $3, $4, $5)
	`,
		e.ID,
		e.Type,
		e.Topic,
		e.Key,
		e.Payload,
	)

	return err
}

func (r *Repository) GetUnprocessed(ctx context.Context, limit int) ([]port.OutboxEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_type, topic, key, payload, created_at
		FROM outbox_events
		WHERE processed_at IS NULL
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []port.OutboxEvent

	for rows.Next() {
		var e port.OutboxEvent

		err := rows.Scan(
			&e.ID,
			&e.Type,
			&e.Topic,
			&e.Key,
			&e.Payload,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, e)
	}

	return result, rows.Err()
}

func (r *Repository) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE outbox_events
		SET processed_at = NOW()
		WHERE id = $1
	`, id)

	return err
}
