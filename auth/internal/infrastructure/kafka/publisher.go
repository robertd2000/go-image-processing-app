package ekafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key []byte, msg any) error {
	var data []byte

	switch v := msg.(type) {
	case []byte:
		data = v
	default:
		var err error
		data, err = json.Marshal(msg)
		if err != nil {
			return err
		}
	}

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: data,
	})

	return err
}
