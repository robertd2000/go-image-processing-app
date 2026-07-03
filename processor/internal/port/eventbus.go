package port

import "context"

type Message struct {
	Key     string
	Value   []byte
	Headers map[string]string
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Close() error
}

type EventHandler func(ctx context.Context, msg Message) error

type EventConsumer interface {
	Consume(ctx context.Context, topics []string, handler EventHandler) error
	Close() error
}
