package model

import (
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
)

func MapToOutput(u *userDomain.User) *UserOutput {
	return &UserOutput{
		ID:       u.ID(),
		Username: u.Username().String(),
		Email:    u.Email().String(),

		Profile: UserProfileOutput{
			Bio:      u.Profile().Bio(),
			Location: u.Profile().Location(),
			Website:  u.Profile().Website(),
		},

		Settings: UserSettingsOutput{
			IsPublic: u.Settings().IsPublic(),
			Theme:    u.Settings().Theme(),
		},
	}
}
