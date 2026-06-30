package image

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
)

type imageService struct {
	imageRepo         imageDomain.Repository
	storage           port.Storage
	metadataExtractor port.Extractor
	txManager         port.TxManager
	outboxRepo        port.OutboxRepository
	jobRepo           port.ProcessingJobRepository
}

type ServiceOption func(*imageService)

func WithOutbox(outboxRepo port.OutboxRepository) ServiceOption {
	return func(s *imageService) {
		s.outboxRepo = outboxRepo
	}
}

func WithJobRepo(jobRepo port.ProcessingJobRepository) ServiceOption {
	return func(s *imageService) {
		s.jobRepo = jobRepo
	}
}

func NewImageService(imageRepo imageDomain.Repository,
	storage port.Storage,
	metadataExtractor port.Extractor,
	txManager port.TxManager,
	opts ...ServiceOption,
) *imageService {
	s := &imageService{
		imageRepo:         imageRepo,
		storage:           storage,
		metadataExtractor: metadataExtractor,
		txManager:         txManager,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *imageService) UploadImage(
	ctx context.Context,
	input model.UploadImageInput,
) (*model.UploadImageOutput, error) {
	if input.UserID == uuid.Nil {
		return nil, imageDomain.ErrInvalidUserID
	}

	if input.Reader == nil {
		return nil, imageDomain.ErrInvalidImageMissingReader
	}

	if input.Size <= 0 {
		return nil, imageDomain.ErrInvalidImageSize
	}

	data, err := readAll(input.Reader, 10<<20)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	size := int64(len(data))

	info, err := s.metadataExtractor.Extract(ctx, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("extract metadata: %w", err)
	}

	meta, err := buildMetadata(info, size)
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

	if err := s.saveImage(ctx, img, data, size, meta.MimeType()); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	return &model.UploadImageOutput{
		ImageID:   img.ID(),
		Status:    string(img.Status()),
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

	if err := s.storage.Put(ctx, key, bytes.NewReader(data), size, mime); err != nil {
		return fmt.Errorf("storage put: %w", err)
	}

	err := s.txManager.WithTx(ctx, func(ctx context.Context, tx txtx.Tx) error {
		if err := s.imageRepo.Save(ctx, tx, img); err != nil {
			return err
		}

		if s.jobRepo != nil {
			now := img.CreatedAt()
			job := &port.ProcessingJob{
				ID:        uuid.New(),
				ImageID:   img.ID(),
				Status:    port.JobStatusPending,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.jobRepo.Create(ctx, tx, job); err != nil {
				return fmt.Errorf("create job: %w", err)
			}
		}

		if s.outboxRepo != nil {
			event := events.NewImageUploaded(
				img.ID(), img.UserID(),
				string(img.StorageKey()), img.OriginalName(),
				img.Metadata().MimeType(),
				img.Metadata().Width(), img.Metadata().Height(),
				img.Metadata().Size(),
			)
			payload, err := json.Marshal(event)
			if err != nil {
				return fmt.Errorf("marshal event: %w", err)
			}
			outboxEvent := &port.OutboxEvent{
				ID:          event.EventID,
				AggregateID: img.ID(),
				EventType:   events.EventTypeImageUploaded,
				Payload:     payload,
				Status:      port.OutboxStatusPending,
				CreatedAt:   event.OccurredAt,
			}

			if err := s.outboxRepo.Save(ctx, tx, outboxEvent); err != nil {
				return fmt.Errorf("save outbox: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		if delErr := s.storage.Delete(ctx, key); delErr != nil {
			return errors.Join(err, fmt.Errorf("storage cleanup: %w", delErr))
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

func (s *imageService) ListImages(ctx context.Context, input model.ListImagesInput) (*model.ListImagesOutput, error) {
	if input.UserID == uuid.Nil {
		return nil, imageDomain.ErrInvalidUserID
	}

	if input.Limit < 0 || input.Offset < 0 {
		return nil, imageDomain.ErrInvalidPagination
	}

	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	images, err := s.imageRepo.GetByUser(ctx, input.UserID, limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("get images: %w", err)
	}

	total, err := s.imageRepo.CountByUser(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("count images: %w", err)
	}

	items := make([]*model.ImageOutput, 0, len(images))

	for _, img := range images {
		url, err := s.storage.GetURL(ctx, string(img.StorageKey()))
		if err != nil {
			return nil, fmt.Errorf("get url: %w", err)
		}

		items = append(items, model.MapToImageOutput(img, url))
	}

	return &model.ListImagesOutput{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: input.Offset,
	}, nil
}

func (s *imageService) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	img, err := s.imageRepo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("get image: %w", err)
	}

	if err = s.imageRepo.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	if err = s.storage.Delete(ctx, string(img.StorageKey())); err != nil {
		return fmt.Errorf("delete image from storage: %w", err)
	}

	return nil
}

func (s *imageService) HandleImageProcessed(ctx context.Context, eventID, imageID uuid.UUID) error {
	if s.jobRepo == nil || s.imageRepo == nil {
		return nil
	}

	ok, err := s.jobRepo.MarkCompleted(ctx, imageID, eventID)
	if err != nil {
		return fmt.Errorf("mark job completed: %w", err)
	}
	if !ok {
		return nil
	}

	return s.imageRepo.UpdateStatus(ctx, imageID, imageDomain.StatusCompleted)
}

func (s *imageService) HandleImageProcessingFailed(ctx context.Context, eventID, imageID uuid.UUID, reason string) error {
	if s.jobRepo == nil || s.imageRepo == nil {
		return nil
	}

	ok, err := s.jobRepo.MarkFailed(ctx, imageID, eventID, reason)
	if err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	if !ok {
		return nil
	}

	return s.imageRepo.UpdateStatus(ctx, imageID, imageDomain.StatusFailed)
}

func detectExtension(mime, filename string) (string, error) {
	switch mime {
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/gif":
		return "gif", nil
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
