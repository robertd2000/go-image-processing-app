package kafkahandler

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type UserSyncService interface {
	UpdateStatus(ctx context.Context, userID uuid.UUID, status userDomain.Status) error
}

type UserStatusChangeHandler struct {
	userSyncSvc UserSyncService
}

func NewUserStatusChangeHandler(userSyncSvc UserSyncService) *UserStatusChangeHandler {
	return &UserStatusChangeHandler{
		userSyncSvc: userSyncSvc,
	}
}

func (s *UserStatusChangeHandler) Handle(ctx context.Context, msg []byte) error {
	var event events.UserStatusUpdatedEvent

	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	status, err := userDomain.ParseStatus(event.Status)
	if err != nil {
		return err
	}

	return s.userSyncSvc.UpdateStatus(ctx, event.ID, status)
}
