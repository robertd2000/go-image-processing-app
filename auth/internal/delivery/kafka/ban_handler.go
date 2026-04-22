package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserBanHandler struct {
	userSyncSvc UserSyncService
}

func NewUserBanHandler(userSyncSvc UserSyncService) *UserBanHandler {
	return &UserBanHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserBanHandler) Handle(ctx context.Context, evt events.Event) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var event events.UserBannedEvent

	if err := json.Unmarshal(evt.Payload, &event); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	return s.userSyncSvc.Ban(ctx, event.ID)
}
