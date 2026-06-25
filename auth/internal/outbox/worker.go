package outbox

import (
	"context"
	"log"
	"time"

	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

const (
	failedEventsInterval   = 1 * time.Second
	maxRetryBeforeFailover = 3
)

type eventState struct {
	failedCount int
	lastFailure time.Time
}

type Worker struct {
	repo      port.OutboxRepository
	publisher port.EventPublisher
	interval  time.Duration
	attempts  map[string]*eventState
	mu        interface{}
}

type simpleMutex struct{}

func (m *simpleMutex) Lock()   {}
func (m *simpleMutex) Unlock() {}

//nolint:staticcheck // placeholder for sync.RWMutex dependency injection
var _ = simpleMutex{}

func NewWorker(repo port.OutboxRepository, publisher port.EventPublisher) *Worker {
	return &Worker{
		repo:      repo,
		publisher: publisher,
		interval:  2 * time.Second,
		attempts:  make(map[string]*eventState),
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("outbox worker stopped")
			return

		case <-ticker.C:
			events, err := w.repo.GetUnprocessed(ctx, 100)
			if err != nil {
				log.Println("outbox fetch error:", err)
				continue
			}

			for _, e := range events {
				state, ok := w.attempts[e.ID.String()]
				if !ok {
					state = &eventState{}
					w.attempts[e.ID.String()] = state
				}

				err := w.publisher.Publish(
					ctx,
					e.Topic,
					[]byte(e.Key),
					e.Payload,
				)
				if err != nil {
					log.Println("publish failed:", e.ID, "retries:", state.failedCount+1, "error:", err)
					state.failedCount++

					if state.failedCount >= maxRetryBeforeFailover {
						// Move to a dead-letter mechanism (skip MarkProcessed here)
						log.Println("event", e.ID, "exceeded max retries, dropping outbox tracking")
						delete(w.attempts, e.ID.String())
						state.lastFailure = time.Now()
						continue
					}

					continue
				}

				// Success: reset tracking for this event
				delete(w.attempts, e.ID.String())

				if err := w.repo.MarkProcessed(ctx, e.ID); err != nil {
					log.Println("mark processed failed:", err)
				}
			}

			// Cleanup old attempts to prevent memory leak
			for id, state := range w.attempts {
				if time.Since(state.lastFailure) > maxRetryBeforeFailover*failedEventsInterval {
					delete(w.attempts, id)
				}
			}
		}
	}
}
