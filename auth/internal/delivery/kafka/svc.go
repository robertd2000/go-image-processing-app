package kafkahandler

import (
	"context"

	"github.com/google/uuid"
)

type UserSyncService interface {
	Delete(ctx context.Context, userID uuid.UUID) error
	Ban(ctx context.Context, userID uuid.UUID) error
	Restore(ctx context.Context, userID uuid.UUID) error
}
