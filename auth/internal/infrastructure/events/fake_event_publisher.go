package eventpub

import (
	"context"
	"sync"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type FakeEventPublisher struct {
	mu sync.Mutex

	Events []events.UserCreatedEvent

	Err error
}

func NewFakeEventPublisher() *FakeEventPublisher {
	return &FakeEventPublisher{
		Events: make([]events.UserCreatedEvent, 0),
	}
}

func (f *FakeEventPublisher) PublishUserCreated(ctx context.Context, e events.UserCreatedEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Err != nil {
		return f.Err
	}

	f.Events = append(f.Events, e)
	return nil
}
