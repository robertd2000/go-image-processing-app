package elog

import (
	"context"
	"fmt"
	"log"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Publish(ctx context.Context, topic string, key []byte, msg any) error {
	event, ok := msg.(events.Event[events.UserCreatedEvent])
	if !ok {
		return fmt.Errorf("unexpected event type: %T", msg)
	}

	e := event.Payload

	log.Printf(
		"[EVENT] UserCreated: user_id=%s email=%s username=%s created_at=%s",
		e.ID, e.Email, e.Username, e.CreatedAt,
	)

	return nil
}
