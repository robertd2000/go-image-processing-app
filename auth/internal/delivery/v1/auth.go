package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/auth/internal/delivery/dao"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/dto"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, username, fistname, lastname, email, password string) error
	Login(ctx context.Context, email string, password string) (*dto.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authHandler struct {
	authSvc AuthService
	logger  *zap.Logger
}

func NewAuthHandler(authSvc AuthService, logger *zap.Logger) *authHandler {
	return &authHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *authHandler) SetupAuthHandler(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.login)
		auth.POST("/register", h.register)
		auth.POST("/logout", h.logout)
		auth.POST("/refresh", h.refresh)
	}
}

// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dao.RegisterRequest true "Register data"
// @Success 200 {string} string
// @Failure 400 {object} dao.ErrorResponse
// @Router /auth/register [post]
func (h *authHandler) register(c *gin.Context) {
	var input dao.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	err := h.authSvc.Register(c.Request.Context(), input.Username, input.Firstname, input.Lastname, input.Email, input.Password)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	c.JSON(http.StatusOK, "signed up")
}

// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dao.LoginRequest true "Login data"
// @Success 200 {object} dao.TokenResponse
// @Failure 400 {object} dao.ErrorResponse
// @Failure 401 {object} dao.ErrorResponse
// @Router /auth/login [post]
func (h *authHandler) login(c *gin.Context) {
	var input dao.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	token, err := h.authSvc.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	response := dao.NewRefreshDAO(token)

	c.JSON(http.StatusOK, response)
}

// @Summary Logout user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dao.RefreshRequest true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *authHandler) logout(c *gin.Context) {
	var input dao.RefreshRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	err := h.authSvc.Logout(c.Request.Context(), input.Token)
	if err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
	}

	c.JSON(http.StatusOK, "logged out")
}

func (h *authHandler) refresh(c *gin.Context) {
	var input dao.RefreshRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	token, err := h.authSvc.Refresh(c.Request.Context(), input.Token)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	response := dao.NewRefreshDAO(token)

	c.JSON(http.StatusOK, response)
}

func mapError(err error) (int, string, string) {
	switch {
	case errors.Is(err, userDomain.ErrWrongCreadentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is wrong"

	case errors.Is(err, userDomain.ErrUserDisabled):
		return http.StatusForbidden, "USER_DISABLED", "user is disabled"

	case errors.Is(err, tokenDomain.ErrInvalidToken):
		return http.StatusUnauthorized, "INVALID_TOKEN", "invalid token"

	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "something went wrong"
	}
}

func respondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, dao.ErrorResponse{
		Error: dao.ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}
