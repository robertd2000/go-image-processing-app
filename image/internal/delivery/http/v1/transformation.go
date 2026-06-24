package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/dao"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	transformDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/transformation"
	transformUsecase "github.com/robertd2000/go-image-processing-app/image/internal/usecase/transformation"
	"go.uber.org/zap"
)

type TransformationService interface {
	RequestTransformation(ctx context.Context, imageID uuid.UUID, spec json.RawMessage) (*transformUsecase.TransformationResult, error)
	GetTransformation(ctx context.Context, id uuid.UUID) (*transformUsecase.TransformationResult, error)
}

type TransformationHandler struct {
	svc    TransformationService
	logger *zap.Logger
}

func NewTransformationHandler(svc TransformationService, logger *zap.Logger) *TransformationHandler {
	return &TransformationHandler{svc: svc, logger: logger}
}

func (h *TransformationHandler) SetupImageTransformRoute(images *gin.RouterGroup) {
	images.POST("/:id/transform", h.requestTransform)
}

func (h *TransformationHandler) SetupTransformationRoutes(api *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	t := api.Group("/transformations")
	t.Use(authMiddleware)
	{
		t.GET("/:id", h.getTransformation)
	}
}

func (h *TransformationHandler) requestTransform(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"})
		return
	}

	var req dao.TransformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.svc.RequestTransformation(c.Request.Context(), imageID, req.Spec)
	if err != nil {
		h.logger.Error("request transformation failed", zap.Error(err))
		if errors.Is(err, imageDomain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, dao.ToTransformResponse(result))
}

func (h *TransformationHandler) getTransformation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transformation id"})
		return
	}

	result, err := h.svc.GetTransformation(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("get transformation failed", zap.Error(err))
		if errors.Is(err, transformDomain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transformation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dao.ToTransformationStatusResponse(result))
}
