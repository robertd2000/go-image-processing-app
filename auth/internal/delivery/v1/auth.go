package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/auth/internal/delivery/dao"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, in model.RegisterInput) error
	Login(ctx context.Context, in model.LoginInput) (*model.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*model.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthHandler struct {
	authSvc AuthService
	logger  *zap.Logger
}

func NewAuthHandler(authSvc AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *AuthHandler) SetupAuthHandler(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.login)
		auth.POST("/register", h.register)
		auth.POST("/logout", h.logout)
		auth.POST("/refresh", h.refresh)
	}
}

// Register godoc
// @Summary Register user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dao.RegisterRequest true "User registration data"
// @Success 201 {object} nil
// @Failure 400 {object} dao.ErrorResponse "Invalid request"
// @Failure 409 {object} dao.ErrorResponse "User already exists"
// @Failure 500 {object} dao.ErrorResponse "Internal error"
// @Router /auth/register [post]
func (h *AuthHandler) register(c *gin.Context) {
	var input dao.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	registerInput := model.RegisterInput{
		Username:  input.Username,
		Email:     input.Email,
		Password:  input.Password,
		FirstName: input.Firstname,
		LastName:  input.Lastname,
	}

	err := h.authSvc.Register(c.Request.Context(), registerInput)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	c.JSON(http.StatusOK, "signed up")
}

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dao.LoginRequest true "Login credentials"
// @Success 200 {object} dao.TokenResponse
// @Failure 400 {object} dao.ErrorResponse "Invalid request"
// @Failure 401 {object} dao.ErrorResponse "Wrong credentials"
// @Failure 500 {object} dao.ErrorResponse "Internal error"
// @Router /auth/login [post]
func (h *AuthHandler) login(c *gin.Context) {
	var input dao.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	loginInput := model.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	}

	token, err := h.authSvc.Login(c.Request.Context(), loginInput)
	if err != nil {
		h.logger.Error("login failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	response := dao.NewRefreshDAO(token)

	c.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary Logout user
// @Description Revokes refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dao.RefreshRequest true "Refresh token"
// @Success 204 {object} nil
// @Failure 400 {object} dao.ErrorResponse "Invalid request"
// @Failure 500 {object} dao.ErrorResponse "Internal error"
// @Router /auth/logout [post]
func (h *AuthHandler) logout(c *gin.Context) {
	var input dao.RefreshRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	err := h.authSvc.Logout(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
	}

	c.JSON(http.StatusOK, "logged out")
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Generates new access and refresh tokens using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dao.RefreshRequest true "Refresh token"
// @Success 200 {object} dao.TokenResponse
// @Failure 400 {object} dao.ErrorResponse "Invalid request"
// @Failure 401 {object} dao.ErrorResponse "Invalid or expired token"
// @Failure 500 {object} dao.ErrorResponse "Internal error"
// @Router /auth/refresh [post]
func (h *AuthHandler) refresh(c *gin.Context) {
	var input dao.RefreshRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	token, err := h.authSvc.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.logger.Error("refresh failed", zap.Error(err))

		status, code, msg := mapError(err)
		respondError(c, status, code, msg)
		return
	}

	response := dao.NewRefreshDAO(token)

	c.JSON(http.StatusOK, response)
}

func mapError(err error) (int, string, string) {
	switch {
	// AUTH
	case errors.Is(err, userDomain.ErrWrongCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is wrong"

	case errors.Is(err, userDomain.ErrUserDisabled):
		return http.StatusForbidden, "USER_DISABLED", "user is disabled"

	case errors.Is(err, tokenDomain.ErrInvalidToken):
		return http.StatusUnauthorized, "INVALID_TOKEN", "invalid token"

	// USER
	case errors.Is(err, userDomain.ErrUserNotFound):
		return http.StatusNotFound, "USER_NOT_FOUND", "user not found"

	case errors.Is(err, userDomain.ErrUserAlreadyExists):
		return http.StatusConflict, "USER_ALREADY_EXISTS", "user already exists"

	// VALIDATION
	case errors.Is(err, userDomain.ErrInvalidEmail):
		return http.StatusBadRequest, "INVALID_EMAIL", "invalid email"

	case errors.Is(err, userDomain.ErrInvalidUsername):
		return http.StatusBadRequest, "INVALID_USERNAME", "invalid username"

	case errors.Is(err, userDomain.ErrInvalidPassword):
		return http.StatusBadRequest, "INVALID_PASSWORD", "invalid password"

	case errors.Is(err, userDomain.ErrInvalidPasswordHash):
		return http.StatusInternalServerError, "INVALID_PASSWORD_HASH", "internal error"

	// ROLES
	case errors.Is(err, userDomain.ErrRoleAlreadyAssigned):
		return http.StatusConflict, "ROLE_ALREADY_ASSIGNED", "role already assigned"

	case errors.Is(err, userDomain.ErrRoleNotAssigned):
		return http.StatusNotFound, "ROLE_NOT_ASSIGNED", "role not assigned"

	// DEFAULT
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
