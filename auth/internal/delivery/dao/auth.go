package dao

import "github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"

type LoginRequest struct {
	Email    string `json:"email" example:"test@mail.com"`
	Password string `json:"password" example:"123456"`
}

type RegisterRequest struct {
	Username  string `json:"username" example:"john"`
	Firstname string `json:"firstname" example:"John"`
	Lastname  string `json:"lastname" example:"Doe"`
	Email     string `json:"email" example:"test@mail.com"`
	Password  string `json:"password" example:"123456"`
}

type RefreshRequest struct {
	Token string `json:"refresh_token" example:"abc123"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"jwt_token"`
	RefreshToken string `json:"refresh_token" example:"refresh_token"`
}

func NewRefreshDAO(accessRefresh *model.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:  accessRefresh.AccessToken,
		RefreshToken: accessRefresh.RefreshToken,
	}
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
