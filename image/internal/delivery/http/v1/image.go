package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/dao"
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/middleware"
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
	svc    ImageService
	logger *zap.Logger
}

func NewImageHandler(svc ImageService, logger *zap.Logger) *ImageHandler {
	return &ImageHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *ImageHandler) SetupImageHandler(api *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	images := api.Group("/images")
	images.Use(authMiddleware)
	{
		images.POST("", h.uploadImage)
		images.GET("", h.listImages)
		images.GET("/:id", h.getImage)
		images.DELETE("/:id", h.deleteImage)
	}
}

// @Summary Upload image
// @Description Upload an image file
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Success 201 {object} dao.UploadImageResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /images [post]
// @Security Bearer
func (h *ImageHandler) uploadImage(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	input := dao.ToUploadImageInput(userID, file, header)

	output, err := h.svc.UploadImage(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("upload image failed", zap.Error(err))
		respondError(c, err)
		return
	}

	img, err := h.svc.GetImage(c.Request.Context(), output.ImageID)
	if err != nil {
		h.logger.Error("get image after upload failed", zap.Error(err))
		respondError(c, err)
		return
	}

	c.JSON(201, dao.ToGetImageResponse(img))
}

// @Summary Get image by ID
// @Description Retrieve image details by ID
// @Tags images
// @Produce json
// @Param id path string true "Image ID (UUID)"
// @Success 200 {object} dao.GetImageResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /images/{id} [get]
// @Security Bearer
func (h *ImageHandler) getImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid image id"})
		return
	}

	img, err := h.svc.GetImage(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("get image failed", zap.Error(err))
		respondError(c, err)
		return
	}

	c.JSON(200, dao.ToGetImageResponse(img))
}

// @Summary List images
// @Description List images with pagination
// @Tags images
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dao.ListImagesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /images [get]
// @Security Bearer
func (h *ImageHandler) listImages(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user"})
		return
	}

	var req dao.ListImagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	input := model.ListImagesInput{
		UserID: userID,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	output, err := h.svc.ListImages(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("list images failed", zap.Error(err))
		respondError(c, err)
		return
	}

	c.JSON(200, dao.ToListImagesResponse(output))
}

// @Summary Delete image
// @Description Delete image by ID
// @Tags images
// @Param id path string true "Image ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /images/{id} [delete]
// @Security Bearer
func (h *ImageHandler) deleteImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid image id"})
		return
	}

	if err := h.svc.DeleteImage(c.Request.Context(), id); err != nil {
		h.logger.Error("delete image failed", zap.Error(err))
		respondError(c, err)
		return
	}

	c.Status(204)
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(string(middleware.ContextUserID))
	if !exists {
		return uuid.Nil, nil
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, nil
	}

	return userID, nil
}
