package kafka

import (
	"context"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/segmentio/kafka-go"
)

type publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers []string) *publisher {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.Hash{},
		Async:    false,
	}
	return &publisher{writer: w}
}

func (p *publisher) Publish(ctx context.Context, topic string, msg port.Message) error {
	kafkaMsg := kafka.Message{
		Topic: topic,
		Key:   []byte(msg.Key),
		Value: msg.Value,
	}
	for k, v := range msg.Headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("kafka publish: %w", err)
	}
	return nil
}

func (p *publisher) Close() error {
	return p.writer.Close()
}
