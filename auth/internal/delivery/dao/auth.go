package dao

import (
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/dto"
)

type UserInput struct {
	Firstname string `json:"firstname" binding:"required" example:"John"`
	Lastname  string `json:"lastname" binding:"required" example:"Doe"`
	Username  string `json:"username" binding:"required" example:"user"`
	Email     string `json:"email" binding:"required,email" example:"user@example.com" format:"email"`
	Password  string `json:"password" binding:"required,min=8" example:"P@ssw0rd" minLength:"8"`
}

type RefreshInput struct {
	Token string `json:"token" binding:"required,min=200" minLength:"200"`
}

type RefreshDAO struct {
	RefreshToken string `json:"access_token" binding:"required,min=200" minLength:"200"`
	AccessToken  string `json:"refresh_token" binding:"required,min=200" minLength:"200"`
}

func NewRefreshDAO(accessRefresh *dto.TokenPair) RefreshDAO {
	return RefreshDAO{
		AccessToken:  accessRefresh.AccessToken,
		RefreshToken: accessRefresh.RefreshToken,
	}
}
