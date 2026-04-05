package dao

import (
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
)

type UpdateUserRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarURL *string `json:"avatar_url"`
}

func (r *UpdateUserRequest) ToInput(userID uuid.UUID) model.UpdateUserInput {
	return model.UpdateUserInput{
		UserID:    userID,
		Username:  r.Username,
		Email:     r.Email,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		AvatarURL: r.AvatarURL,
	}
}

type UpdateProfileRequest struct {
	Bio      *string    `json:"bio"`
	Location *string    `json:"location"`
	Website  *string    `json:"website"`
	Birthday *time.Time `json:"birthday"`
}

type UpdateSettingsRequest struct {
	IsPublic           *bool   `json:"is_public"`
	AllowNotifications *bool   `json:"allow_notifications"`
	Theme              *string `json:"theme"`
}
