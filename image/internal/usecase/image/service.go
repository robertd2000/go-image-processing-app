package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
)

type imageService struct {
	imageRepo         imageDomain.Repository
	storage           port.Storage
	metadataExtractor port.Extractor
}

func NewImageService(imageRepo imageDomain.Repository,
	storage port.Storage,
	metadataExtractor port.Extractor,
) *imageService {
	return &imageService{
		imageRepo:         imageRepo,
		storage:           storage,
		metadataExtractor: metadataExtractor,
	}
}

func (s *imageService) UploadImage(
	ctx context.Context,
	input model.UploadImageInput,
) (*model.UploadImageOutput, error) {
	// --- validation ---
	if input.UserID == uuid.Nil {
		return nil, imageDomain.ErrInvalidUserID
	}

	if input.Reader == nil {
		return nil, imageDomain.ErrInvalidImageMissingReader
	}

	if input.Size <= 0 {
		return nil, imageDomain.ErrInvalidImageSize
	}

	// --- read full file ---
	data, err := io.ReadAll(input.Reader)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	size := int64(len(data))

	// --- extract metadata (width/height/mime) ---
	reader := bytes.NewReader(data)

	info, err := s.metadataExtractor.Extract(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("extract metadata: %w", err)
	}

	meta, err := imageDomain.NewImageMetadata(
		info.Width,
		info.Height,
		size,
		info.MimeType,
	)
	if err != nil {
		return nil, fmt.Errorf("create metadata: %w", err)
	}

	ext, err := detectExtension(info.MimeType, input.Filename)
	if err != nil {
		return nil, fmt.Errorf("detect extension: %w", err)
	}

	img, err := imageDomain.NewImage(
		input.UserID,
		input.Filename,
		meta,
		ext,
	)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	// --- upload to storage ---
	err = s.storage.Put(
		ctx,
		string(img.StorageKey()),
		bytes.NewReader(data),
		size,
		meta.MimeType,
	)
	if err != nil {
		return nil, fmt.Errorf("storage put: %w", err)
	}

	err = s.imageRepo.Save(ctx, img)
	if err != nil {
		// rollback storage
		_ = s.storage.Delete(ctx, string(img.StorageKey()))
		return nil, fmt.Errorf("save image: %w", err)
	}

	return &model.UploadImageOutput{
		ImageID:   img.ID(),
		CreatedAt: img.CreatedAt(),
	}, nil
}

func detectExtension(mime, filename string) (string, error) {
	switch mime {
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/webp":
		return "webp", nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	ext = strings.TrimPrefix(ext, ".")

	if ext == "" {
		return "", fmt.Errorf("cannot detect extension")
	}

	return ext, nil
}
