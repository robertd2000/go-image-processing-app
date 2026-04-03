package elog

import (
	"context"
	"log"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) PublishUserCreated(ctx context.Context, e events.UserCreatedEvent) error {
	log.Printf(
		"[EVENT] UserCreated: user_id=%s email=%s username=%s created_at=%s",
		e.ID, e.Email, e.Username, e.CreatedAt,
	)
	return nil
}
