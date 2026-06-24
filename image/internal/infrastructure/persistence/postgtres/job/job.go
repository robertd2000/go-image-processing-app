package jobpg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"go.uber.org/zap"
)

type jobRepository struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewJobRepository(pool *pgxpool.Pool, logger *zap.SugaredLogger) *jobRepository {
	return &jobRepository{pool: pool, logger: logger}
}

func (r *jobRepository) Create(ctx context.Context, tx txtx.Tx, job *port.ProcessingJob) error {
	query := `
		INSERT INTO image_processing_jobs (id, image_id, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5)
	`
	return tx.Exec(ctx, query, job.ID, job.ImageID, string(job.Status), job.CreatedAt, job.UpdatedAt)
}

func (r *jobRepository) MarkCompleted(ctx context.Context, imageID, eventID uuid.UUID) (bool, error) {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE image_processing_jobs
		SET status = 'completed', event_id = $1, updated_at = NOW()
		WHERE image_id = $2 AND status NOT IN ('completed', 'failed')
	`, eventID, imageID)
	if err != nil {
		return false, fmt.Errorf("mark job completed: %w", err)
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *jobRepository) MarkFailed(ctx context.Context, imageID uuid.UUID, eventID uuid.UUID, reason string) (bool, error) {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE image_processing_jobs
		SET status = 'failed', event_id = $1, error_message = $2, updated_at = NOW()
		WHERE image_id = $3 AND status NOT IN ('completed', 'failed')
	`, eventID, reason, imageID)
	if err != nil {
		return false, fmt.Errorf("mark job failed: %w", err)
	}
	return cmd.RowsAffected() > 0, nil
}
