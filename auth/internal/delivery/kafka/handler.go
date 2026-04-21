// Package kafkahandler
package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	domainevents "github.com/robertd2000/go-image-processing-app/auth/internal/domain/events"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserSyncService interface {
	Delete(ctx context.Context, userID uuid.UUID) error
}

type UserDeletedHandler struct {
	userSyncSvc UserSyncService
}

func NewUserDeletedHandler(userSyncSvc UserSyncService) *UserDeletedHandler {
	return &UserDeletedHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserDeletedHandler) Handle(ctx context.Context, evt events.Event) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var event domainevents.UserDeletedEvent

	if err := json.Unmarshal(evt.Payload, &event); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	return s.userSyncSvc.Delete(ctx, event.ID)
}
