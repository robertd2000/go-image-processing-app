package port

import (
	"context"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key []byte, msg any) error
}
