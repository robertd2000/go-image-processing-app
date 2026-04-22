package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserRestoreHandler struct {
	userSyncSvc UserSyncService
}

func NewUserRestoreHandler(userSyncSvc UserSyncService) *UserRestoreHandler {
	return &UserRestoreHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserRestoreHandler) Handle(ctx context.Context, evt events.Event) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var event events.UserRestoredEvent

	if err := json.Unmarshal(evt.Payload, &event); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	return s.userSyncSvc.Restore(ctx, event.ID)
}
