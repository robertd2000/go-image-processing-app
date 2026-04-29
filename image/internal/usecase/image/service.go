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
	txManager         port.TxManager
}

func NewImageService(imageRepo imageDomain.Repository,
	storage port.Storage,
	metadataExtractor port.Extractor,
	txManager port.TxManager,
) *imageService {
	return &imageService{
		imageRepo:         imageRepo,
		storage:           storage,
		metadataExtractor: metadataExtractor,
		txManager:         txManager,
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

	// --- read file (with limit) ---
	data, err := readAll(input.Reader, 10<<20) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	size := int64(len(data))

	// --- extract metadata ---
	info, err := s.metadataExtractor.Extract(ctx, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("extract metadata: %w", err)
	}

	// --- build domain metadata ---
	meta, err := buildMetadata(info, size)
	if err != nil {
		return nil, fmt.Errorf("create metadata: %w", err)
	}

	// --- extension ---
	ext, err := detectExtension(info.MimeType, input.Filename)
	if err != nil {
		return nil, fmt.Errorf("detect extension: %w", err)
	}

	// --- domain image ---
	img, err := imageDomain.NewImage(
		input.UserID,
		input.Filename,
		meta,
		ext,
	)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	// --- persist ---
	if err := s.saveImage(ctx, img, data, size, meta.MimeType()); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	return &model.UploadImageOutput{
		ImageID:   img.ID(),
		CreatedAt: img.CreatedAt(),
	}, nil
}

func (s *imageService) saveImage(
	ctx context.Context,
	img *imageDomain.Image,
	data []byte,
	size int64,
	mime string,
) error {
	key := string(img.StorageKey())

	if err := s.storage.Put(
		ctx,
		key,
		bytes.NewReader(data),
		size,
		mime,
	); err != nil {
		return fmt.Errorf("storage put: %w", err)
	}

	err := s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {
		return s.imageRepo.Save(ctx, tx, img)
	})

	if err != nil {
		if delErr := s.storage.Delete(ctx, key); delErr != nil {
			return fmt.Errorf("db save failed: %w; cleanup failed: %v", err, delErr)
		}

		return err
	}

	return nil
}

func (s *imageService) GetImage(ctx context.Context, imageID uuid.UUID) (*model.ImageOutput, error) {
	img, err := s.imageRepo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}

	url, err := s.storage.GetURL(ctx, string(img.StorageKey()))
	if err != nil {
		return nil, fmt.Errorf("get url: %w", err)
	}

	return model.MapToImageOutput(img, url), nil
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

func readAll(r io.Reader, maxSize int64) ([]byte, error) {
	lr := io.LimitReader(r, maxSize)

	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) >= maxSize {
		return nil, fmt.Errorf("file too large")
	}

	return data, nil
}

func buildMetadata(
	info port.ImageInfo,
	size int64,
) (imageDomain.ImageMetadata, error) {
	return imageDomain.NewImageMetadata(
		info.Width,
		info.Height,
		size,
		info.MimeType,
	)
}
