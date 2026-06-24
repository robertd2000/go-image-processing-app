package events

import (
	"context"
	"log"
	"time"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

func RunOutboxRelay(
	ctx context.Context,
	outboxRepo port.OutboxRepository,
	publisher port.EventPublisher,
	topic string,
	interval time.Duration,
	batchSize int,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("outbox relay stopped")
			return
		case <-ticker.C:
			pending, err := outboxRepo.FetchPending(ctx, batchSize)
			if err != nil {
				log.Printf("outbox relay: fetch failed: %v", err)
				continue
			}
			for _, event := range pending {
				msg := port.Message{
					Key:   event.AggregateID.String(),
					Value: event.Payload,
					Headers: map[string]string{
						"event_type": event.EventType,
					},
				}
				if err := publisher.Publish(ctx, topic, msg); err != nil {
					log.Printf("outbox relay: publish failed for event %s: %v", event.ID, err)
					_ = outboxRepo.MarkFailed(ctx, event.ID)
					continue
				}
				if err := outboxRepo.MarkPublished(ctx, event.ID); err != nil {
					log.Printf("outbox relay: mark published failed for event %s: %v", event.ID, err)
				}
			}
		}
	}
}
