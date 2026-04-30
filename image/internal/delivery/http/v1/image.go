package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
	"go.uber.org/zap"
)

type ImageService interface {
	UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error)
	GetImage(ctx context.Context, imageID uuid.UUID) (*model.ImageOutput, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
	ListImages(ctx context.Context, input model.ListImagesInput) (*model.ListImagesOutput, error)
}

type ImageHandler struct {
	userSvc ImageService
	logger  *zap.Logger
}

func NewUserHandler(userSvc ImageService, logger *zap.Logger) *ImageHandler {
	return &ImageHandler{
		userSvc: userSvc,
		logger:  logger,
	}
}

func (h *ImageHandler) SetupImageHandler(api *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
}
