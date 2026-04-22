package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"

	domainevents "github.com/robertd2000/go-image-processing-app/user/internal/domain/events"
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

func (h *UserCreatedHandler) Handle(ctx context.Context, evt events.Event) error {
	if evt.Version != 1 {
		return fmt.Errorf("unsupported version: %d", evt.Version)
	}

	var payload domainevents.UserCreatedEvent

	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	input := model.CreateUserInput{
		ID:       payload.ID,
		Username: payload.Username,
		Email:    payload.Email,
	}

	return h.userService.CreateFromEvent(ctx, input)
}
