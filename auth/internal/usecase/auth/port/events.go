package port

import "github.com/google/uuid"

type UserCreatedEvent struct {
	UserID    uuid.UUID
	Email     string
	FirstName string
	LastName  string
}
