package port

import "context"

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, event UserCreatedEvent) error
}
