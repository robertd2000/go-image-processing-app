package port

import (
	"context"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, event events.UserCreatedEvent) error
}
