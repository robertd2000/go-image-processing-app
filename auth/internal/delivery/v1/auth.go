package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/dto"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, username, fistname, lastname, email, password string) error
	Login(ctx context.Context, email string, password string) (*dto.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthHandler interface {
	// signIn(c *gin.Context)
	// signUp(c *gin.Context)
	// signOut(c *gin.Context)
	// refreshToken(c *gin.Context)
	SetupAuthHandler(api *gin.RouterGroup)
}

type authHandler struct {
	authSvc AuthService
	logger  *zap.Logger
}

func NewAuthHandler(authSvc AuthService, logger *zap.Logger) AuthHandler {
	return &authHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *authHandler) SetupAuthHandler(api *gin.RouterGroup) {
	// auth := api.Group("/auth")
	{
		// auth.POST("/sign-in", h.signIn)
		// auth.POST("/sign-up", h.signUp)
		// auth.POST("/sign-out", h.signOut)
		// auth.POST("/refresh", h.refreshToken)
	}
}
