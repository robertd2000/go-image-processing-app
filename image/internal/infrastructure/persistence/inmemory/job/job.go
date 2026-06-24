package jobmem

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type jobRepo struct {
	mu    sync.Mutex
	data  map[uuid.UUID]*port.ProcessingJob
	clock func() time.Time
}

func NewInMemoryJobRepo() *jobRepo {
	return &jobRepo{
		data:  make(map[uuid.UUID]*port.ProcessingJob),
		clock: time.Now,
	}
}

func (r *jobRepo) Create(ctx context.Context, tx txtx.Tx, job *port.ProcessingJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *job
	clone.CreatedAt = r.clock()
	clone.UpdatedAt = clone.CreatedAt
	r.data[clone.ID] = &clone
	return nil
}

func (r *jobRepo) MarkCompleted(ctx context.Context, imageID, eventID uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, j := range r.data {
		if j.ImageID == imageID {
			if j.Status == port.JobStatusCompleted || j.Status == port.JobStatusFailed {
				return false, nil
			}
			j.Status = port.JobStatusCompleted
			evID := eventID
			j.EventID = &evID
			j.UpdatedAt = r.clock()
			return true, nil
		}
	}
	return false, nil
}

func (r *jobRepo) MarkFailed(ctx context.Context, imageID uuid.UUID, eventID uuid.UUID, reason string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, j := range r.data {
		if j.ImageID == imageID {
			if j.Status == port.JobStatusCompleted || j.Status == port.JobStatusFailed {
				return false, nil
			}
			j.Status = port.JobStatusFailed
			evID := eventID
			j.EventID = &evID
			j.ErrorMessage = reason
			j.UpdatedAt = r.clock()
			return true, nil
		}
	}
	return false, nil
}
