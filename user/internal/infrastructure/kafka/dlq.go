package ckafka

import "context"

type DLQProducer interface {
	Publish(ctx context.Context, topic string, key []byte, msg []byte) error
}
