package eventsmem

import (
	"context"
	"sync"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type FakePublisher struct {
	mu       sync.Mutex
	Messages []port.Message
	Topics   []string
}

func NewFakePublisher() *FakePublisher {
	return &FakePublisher{}
}

func (p *FakePublisher) Publish(_ context.Context, topic string, msg port.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages = append(p.Messages, msg)
	p.Topics = append(p.Topics, topic)
	return nil
}

func (p *FakePublisher) Close() error { return nil }

type FakeConsumer struct {
	mu      sync.Mutex
	handler port.EventHandler
}

func NewFakeConsumer() *FakeConsumer {
	return &FakeConsumer{}
}

func (c *FakeConsumer) Consume(_ context.Context, _ []string, handler port.EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handler = handler
	return nil
}

func (c *FakeConsumer) Close() error { return nil }

func (c *FakeConsumer) Trigger(ctx context.Context, msg port.Message) error {
	c.mu.Lock()
	handler := c.handler
	c.mu.Unlock()
	if handler == nil {
		return nil
	}
	return handler(ctx, msg)
}
