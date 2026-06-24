package events

import (
	"context"
	"log"
	"maps"
	"time"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

// WithDLQ wraps an EventHandler so that when it returns an error, the original
// message is published to a dead-letter topic with error metadata.
func WithDLQ(handler port.EventHandler, publisher port.EventPublisher, dlqTopic string) port.EventHandler {
	return func(ctx context.Context, msg port.Message) error {
		err := handler(ctx, msg)
		if err == nil {
			return nil
		}

		headers := make(map[string]string, len(msg.Headers)+2)
		maps.Copy(headers, msg.Headers)
		headers["dlq_error"] = err.Error()
		headers["dlq_timestamp"] = time.Now().UTC().Format(time.RFC3339)

		dlqMsg := port.Message{
			Key:     msg.Key,
			Value:   msg.Value,
			Headers: headers,
		}
		if pubErr := publisher.Publish(ctx, dlqTopic, dlqMsg); pubErr != nil {
			log.Printf("DLQ: publish failed: %v", pubErr)
		}

		return nil
	}
}
