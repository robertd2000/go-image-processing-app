package kafkahandler

import (
	"context"
	"encoding/json"

	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
	"github.com/robertd2000/go-image-processing-app/user/pkg/events"
)

type UserService interface {
	CreateFromEvent(ctx context.Context, input model.CreateUserInput) error
}

type UserCreatedHandler struct {
	userService UserService
}

func NewUserCreatedHandler(s UserService) *UserCreatedHandler {
	return &UserCreatedHandler{userService: s}
}

func (h *UserCreatedHandler) Handle(ctx context.Context, msg []byte) error {
	var event events.Event[events.UserCreatedEvent]

	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	input := model.CreateUserInput{
		ID:       event.Payload.ID,
		Username: event.Payload.Username,
		Email:    event.Payload.Email,
	}

	return h.userService.CreateFromEvent(ctx, input)
}
