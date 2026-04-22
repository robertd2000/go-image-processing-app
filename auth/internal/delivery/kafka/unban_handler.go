package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserUnbanHandler struct {
	userSyncSvc UserSyncService
}

func NewUserUnbanHandler(userSyncSvc UserSyncService) *UserUnbanHandler {
	return &UserUnbanHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserUnbanHandler) Handle(ctx context.Context, evt events.Event) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var event events.UserUnbannedEvent

	if err := json.Unmarshal(evt.Payload, &event); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	return s.userSyncSvc.Unban(ctx, event.ID)
}
