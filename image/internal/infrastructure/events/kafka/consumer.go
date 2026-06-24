package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/segmentio/kafka-go"
)

type consumer struct {
	brokers []string
	groupID string
	reader  *kafka.Reader
}

func NewConsumer(brokers []string, groupID string) *consumer {
	return &consumer{brokers: brokers, groupID: groupID}
}

func (c *consumer) Consume(ctx context.Context, topics []string, handler port.EventHandler) error {
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.groupID,
		GroupTopics: topics,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer c.reader.Close()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("kafka consume: %w", err)
		}

		headers := make(map[string]string)
		for _, h := range msg.Headers {
			headers[h.Key] = string(h.Value)
		}

		if err := handler(ctx, port.Message{
			Key:     string(msg.Key),
			Value:   msg.Value,
			Headers: headers,
		}); err != nil {
			log.Printf("kafka consumer: handler error: %v", err)
		}
	}
}

func (c *consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
