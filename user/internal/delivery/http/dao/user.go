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

func (r *UpdateProfileRequest) ToInput(userID uuid.UUID) model.UpdateProfileInput {
	return model.UpdateProfileInput{
		UserID:   userID,
		Bio:      r.Bio,
		Location: r.Location,
		Website:  r.Website,
		Birthday: r.Birthday,
	}
}

type UpdateSettingsRequest struct {
	IsPublic           *bool   `json:"is_public"`
	AllowNotifications *bool   `json:"allow_notifications"`
	Theme              *string `json:"theme"`
}

func (r *UpdateSettingsRequest) ToInput(userID uuid.UUID) model.UpdateSettingsInput {
	return model.UpdateSettingsInput{
		UserID:             userID,
		IsPublic:           r.IsPublic,
		AllowNotifications: r.AllowNotifications,
		Theme:              r.Theme,
	}
}

// UserOutput represents full user data
// @Description User with profile and settings
type UserOutput struct {
	ID       uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string    `json:"username" example:"john_doe"`
	Email    string    `json:"email" example:"john@mail.com"`

	Profile  UserProfileOutput  `json:"profile"`
	Settings UserSettingsOutput `json:"settings"`
}

// UserProfileOutput represents user profile
// @Description Additional profile information
type UserProfileOutput struct {
	Bio      *string `json:"bio" example:"Software engineer"`
	Location *string `json:"location" example:"Berlin"`
	Website  *string `json:"website" example:"https://example.com"`
}

// UserSettingsOutput represents user settings
// @Description User preferences and visibility settings
type UserSettingsOutput struct {
	IsPublic bool   `json:"is_public" example:"true"`
	Theme    string `json:"theme" example:"dark"`
}

type UsersListResponse struct {
	Items []UserOutput `json:"items"`
	Total int          `json:"total" example:"100"`
}
