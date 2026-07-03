package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProcessingJobStatus string

const (
	JobStatusPending   ProcessingJobStatus = "pending"
	JobStatusPublished ProcessingJobStatus = "published"
	JobStatusCompleted ProcessingJobStatus = "completed"
	JobStatusFailed    ProcessingJobStatus = "failed"
)

type ProcessingJob struct {
	ID           uuid.UUID
	ImageID      uuid.UUID
	Status       ProcessingJobStatus
	EventID      *uuid.UUID
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProcessingJobRepository interface {
	Create(ctx context.Context, tx Tx, job *ProcessingJob) error
	MarkCompleted(ctx context.Context, imageID, eventID uuid.UUID) (bool, error)
	MarkFailed(ctx context.Context, imageID uuid.UUID, eventID uuid.UUID, reason string) (bool, error)
}
