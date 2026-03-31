package elog

import (
	"context"
	"log"

	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/port"
)

type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) PublishUserCreated(ctx context.Context, e port.UserCreatedEvent) error {
	log.Printf(
		"[EVENT] UserCreated: user_id=%s email=%s first_name=%s last_name=%s",
		e.UserID,
		e.Email,
		e.FirstName,
		e.LastName,
	)
	return nil
}
