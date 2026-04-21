package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type Handler interface {
	Handle(ctx context.Context, evt events.RawEvent) error
}

type Middleware func(Handler) Handler

type Dispatcher struct {
	handlers   map[string]Handler
	middleware []Middleware
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

func (d *Dispatcher) Register(eventType string, h Handler) {
	d.handlers[eventType] = h
}

func (d *Dispatcher) Use(mw Middleware) {
	d.middleware = append(d.middleware, mw)
}

func (d *Dispatcher) wrap(h Handler) Handler {
	for i := len(d.middleware) - 1; i >= 0; i-- {
		h = d.middleware[i](h)
	}
	return h
}

func (d *Dispatcher) Dispatch(ctx context.Context, msg []byte) error {
	var evt events.RawEvent

	if err := json.Unmarshal(msg, &evt); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	log.Println("EVENT TYPE:", evt.EventType)

	h, ok := d.handlers[evt.EventType]
	if !ok {
		log.Println("NO HANDLER FOR:", evt.EventType)
		return nil
	}

	h = d.wrap(h)

	return h.Handle(ctx, evt)
}
