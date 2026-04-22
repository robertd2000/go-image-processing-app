package ckafka

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/robertd2000/go-image-processing-app/user/pkg/events"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	dlq    DLQProducer
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0,
		}),
	}
}

func (c *Consumer) Start(ctx context.Context, handler func(context.Context, []byte) error) {
	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer stopped")
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("kafka consumer context canceled")
				return
			}

			log.Println("kafka fetch error:", err)
			continue
		}

		// retry
		err = c.handleWithRetry(ctx, handler, msg.Value)
		if err != nil {
			log.Println("message failed after retries:", err)

			_ = c.dlq.Publish(ctx, events.UserEventsDLQ, msg.Key, msg.Value)

			_ = c.reader.CommitMessages(ctx, msg)

			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Println("commit error:", err)
		}
	}
}

func (c *Consumer) handleWithRetry(
	ctx context.Context,
	handler func(context.Context, []byte) error,
	msg []byte,
) error {
	const maxRetries = 3

	var err error

	for i := range maxRetries {
		err = handler(ctx, msg)
		if err == nil {
			return nil
		}

		log.Printf("handler failed (attempt %d): %v\n", i+1, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}

	return err
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
