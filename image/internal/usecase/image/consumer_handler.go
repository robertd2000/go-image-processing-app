package image

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

func NewProcessingResultHandler(svc *imageService) port.EventHandler {
	return func(ctx context.Context, msg port.Message) error {
		eventType := msg.Headers["event_type"]

		switch eventType {
		case events.EventTypeImageProcessingCompleted:
			var ev events.ImageProcessingCompleted
			if err := json.Unmarshal(msg.Value, &ev); err != nil {
				return fmt.Errorf("unmarshal completed event: %w", err)
			}
			return svc.HandleImageProcessed(ctx, ev.EventID, ev.ImageID)

		case events.EventTypeImageProcessingFailed:
			var ev events.ImageProcessingFailed
			if err := json.Unmarshal(msg.Value, &ev); err != nil {
				return fmt.Errorf("unmarshal failed event: %w", err)
			}
			return svc.HandleImageProcessingFailed(ctx, ev.EventID, ev.ImageID, ev.Reason)

		default:
			log.Printf("unknown event type: %s", eventType)
			return nil
		}
	}
}
