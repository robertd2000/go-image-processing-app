package ekafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,

			MinBytes: 1e3,
			MaxBytes: 1e6,
		}),
	}
}

func (c *Consumer) Start(ctx context.Context, handler func(context.Context, []byte) error) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		if err := handler(ctx, m.Value); err != nil {
			log.Printf("failed to handle message: %v", err)
			continue
		}
	}
}

func (c *Consumer) Close() {
	c.reader.Close()
}
