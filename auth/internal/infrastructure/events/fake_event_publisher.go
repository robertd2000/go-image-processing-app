package eventpub

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type FakeEventPublisher struct {
	mu sync.Mutex

	Events []events.Event[events.UserCreatedEvent]

	Err error
}

func NewFakeEventPublisher() *FakeEventPublisher {
	return &FakeEventPublisher{
		Events: make([]events.Event[events.UserCreatedEvent], 0),
	}
}

func (f *FakeEventPublisher) Publish(ctx context.Context, topic string, key []byte, msg any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Err != nil {
		return f.Err
	}

	event, ok := msg.(events.Event[events.UserCreatedEvent])
	if !ok {
		return fmt.Errorf("unexpected event type: %T", msg)
	}

	f.Events = append(f.Events, event)
	return nil
}
