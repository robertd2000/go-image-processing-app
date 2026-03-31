package port

import (
	"context"
	"sync"
)

type FakeEventPublisher struct {
	mu sync.Mutex

	Events []UserCreatedEvent

	Err error
}

func NewFakeEventPublisher() *FakeEventPublisher {
	return &FakeEventPublisher{
		Events: make([]UserCreatedEvent, 0),
	}
}

func (f *FakeEventPublisher) PublishUserCreated(ctx context.Context, e UserCreatedEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Err != nil {
		return f.Err
	}

	f.Events = append(f.Events, e)
	return nil
}
