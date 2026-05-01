package dao

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
)

type UploadImageResponse struct {
	ImageID  string `json:"image_id"`
	FileName string `json:"file_name"`
	URL      string `json:"url,omitempty"`

	Width  int `json:"width"`
	Height int `json:"height"`

	Size int64 `json:"size"`

	MimeType string `json:"mime_type"`

	CreatedAt string `json:"created_at"`
}

type GetImageResponse struct {
	ImageID string `json:"image_id"`
	UserID  string `json:"user_id"`

	FileName string `json:"file_name"`

	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`

	Width  int `json:"width"`
	Height int `json:"height"`

	URL string `json:"url"`

	CreatedAt string `json:"created_at"`
}

type ListImagesRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

type ListImagesResponse struct {
	Items  []GetImageResponse `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

func ToGetImageResponse(
	img *model.ImageOutput,
) GetImageResponse {
	return GetImageResponse{
		ImageID: img.ImageID.String(),
		UserID:  img.UserID.String(),

		FileName: img.FileName,

		MimeType: img.MimeType,
		Size:     img.Size,

		Width:  img.Width,
		Height: img.Height,

		URL: img.URL,

		CreatedAt: img.CreatedAt.Format(time.RFC3339),
	}
}

func ToUploadImageInput(
	userID uuid.UUID,
	file multipart.File,
	header *multipart.FileHeader,
) model.UploadImageInput {

	return model.UploadImageInput{
		UserID:      userID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	}
}

func ToListImagesResponse(
	output *model.ListImagesOutput,
) ListImagesResponse {

	items := make([]GetImageResponse, 0, len(output.Items))

	for _, img := range output.Items {
		items = append(items, ToGetImageResponse(img))
	}

	return ListImagesResponse{
		Items:  items,
		Total:  output.Total,
		Limit:  output.Limit,
		Offset: output.Offset,
	}
}
