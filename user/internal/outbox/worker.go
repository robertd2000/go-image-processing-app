package outbox

import (
	"context"
	"log"
	"time"

	"github.com/robertd2000/go-image-processing-app/user/internal/port"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, key []byte, msg any) error
}

type Worker struct {
	repo      port.OutboxRepository
	publisher Publisher
	interval  time.Duration
}

func NewWorker(repo port.OutboxRepository, publisher Publisher) *Worker {
	return &Worker{
		repo:      repo,
		publisher: publisher,
		interval:  2 * time.Second,
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
				log.Println("OUTBOX: publishing event", e.ID, e.Topic)
				err := w.publisher.Publish(
					ctx,
					e.Topic,
					[]byte(e.Key),
					e.Payload,
				)
				log.Println("OUTBOX: published", e.ID)
				if err != nil {
					log.Println("publish failed:", err)
					continue
				}

				if err := w.repo.MarkProcessed(ctx, e.ID); err != nil {
					log.Println("mark processed failed:", err)
				}
			}
		}
	}
}
