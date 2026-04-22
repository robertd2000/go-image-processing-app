package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/user/pkg/events"
)

type Handler interface {
	Handle(ctx context.Context, evt events.Event) error
}

type Dispatcher struct {
	handlers map[string]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

func (d *Dispatcher) Register(eventType string, h Handler) {
	d.handlers[eventType] = h
}

func (d *Dispatcher) Dispatch(ctx context.Context, msg []byte) error {
	var evt events.Event

	if err := json.Unmarshal(msg, &evt); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	h, ok := d.handlers[evt.EventType]
	if !ok {
		return nil
	}

	return h.Handle(ctx, evt)
}
