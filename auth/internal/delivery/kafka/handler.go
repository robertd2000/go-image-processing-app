package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserSyncService interface {
	UpdateStatus(ctx context.Context, userID uuid.UUID, status userDomain.Status) error
}

type UserStatusUpdatedHandler struct {
	userSyncSvc UserSyncService
}

func NewUserStatusUpdatedHandler(userSyncSvc UserSyncService) *UserStatusUpdatedHandler {
	return &UserStatusUpdatedHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserStatusUpdatedHandler) Handle(ctx context.Context, evt events.RawEvent) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var event events.UserStatusUpdatedEvent

	if err := json.Unmarshal(evt.Payload, &event); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	status, err := userDomain.ParseStatus(event.Status)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return s.userSyncSvc.UpdateStatus(ctx, event.ID, status)
}
