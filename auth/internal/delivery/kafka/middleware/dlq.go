package kafkamiddleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
	"github.com/segmentio/kafka-go"
)

type DLQProducer struct {
	writer *kafka.Writer
}

func NewDLQProducer(brokers []string, topic string) *DLQProducer {
	return &DLQProducer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
		},
	}
}

func (p *DLQProducer) Send(ctx context.Context, evt events.Event, reason error) error {
	body, _ := json.Marshal(map[string]interface{}{
		"event": evt,
		"error": reason.Error(),
		"time":  time.Now(),
	})

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: body,
	})
}
