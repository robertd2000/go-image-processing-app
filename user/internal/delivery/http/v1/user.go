package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/user/internal/delivery/http/dao"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user/model"
	"go.uber.org/zap"
)

type UserService interface {
	CreateFromEvent(ctx context.Context, input model.CreateUserInput) error
	Update(ctx context.Context, input model.UpdateUserInput) error
	UpdateProfile(ctx context.Context, input model.UpdateProfileInput) error
	UpdateSettings(ctx context.Context, input model.UpdateSettingsInput) error
	Delete(ctx context.Context, userID uuid.UUID) error
	GetByID(ctx context.Context, userID uuid.UUID) (*model.UserOutput, error)
	GetByEmail(ctx context.Context, email string) (*model.UserOutput, error)
	List(ctx context.Context, filter model.UserFilterInput) ([]*model.UserOutput, error)
	Count(ctx context.Context, filter model.UserFilterInput) (int, error)
}

type UserHandler struct {
	userSvc UserService
	logger  *zap.Logger
}

func NewUserHandler(userSvc UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		logger:  logger,
	}
}

func (h *UserHandler) SetupUserHandler(api *gin.RouterGroup) {
	user := api.Group("/users")
	{
		user.POST("/", h.createUser)
		user.PUT("/:id", h.updateUser)
		user.PUT("/:id/profile", h.updateProfile)
		user.PUT("/:id/settings", h.updateSettings)
		user.DELETE("/:id", h.deleteUser)
		user.GET("/:id", h.getUserByID)
		user.GET("/email/:email", h.getUserByEmail)
		user.GET("/", h.listUsers)
	}
}

func (h *UserHandler) createUser(c *gin.Context) {
	var input model.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := h.userSvc.CreateFromEvent(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to create user", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(201, gin.H{"message": "user created"})
}

func (h *UserHandler) updateUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	var req dao.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	input := req.ToInput(userID)
	err = h.userSvc.Update(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"message": "user updated"})
}

func (h *UserHandler) updateProfile(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}
	var req dao.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	input := req.ToInput(userID)

	err = h.userSvc.UpdateProfile(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to update profile", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.Status(204)
}

func (h *UserHandler) updateSettings(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}
	var req dao.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	input := req.ToInput(userID)
	err = h.userSvc.UpdateSettings(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to update settings", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"message": "settings updated"})
}

func (h *UserHandler) deleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}
	err = h.userSvc.Delete(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"message": "user deleted"})
}

func (h *UserHandler) getUserByID(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, user)
}

func (h *UserHandler) getUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := h.userSvc.GetByEmail(c.Request.Context(), email)
	if err != nil {
		h.logger.Error("failed to get user by email", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, user)
}

func (h *UserHandler) listUsers(c *gin.Context) {
	var filter model.UserFilterInput
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	users, err := h.userSvc.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, users)
}
