package transformation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	transformDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type Service struct {
	imageRepo     imageDomain.Repository
	transformRepo transformDomain.Repository
	txManager     port.TxManager
	outboxRepo    port.OutboxRepository
}

func NewService(
	imageRepo imageDomain.Repository,
	transformRepo transformDomain.Repository,
	txManager port.TxManager,
	outboxRepo port.OutboxRepository,
) *Service {
	return &Service{
		imageRepo:     imageRepo,
		transformRepo: transformRepo,
		txManager:     txManager,
		outboxRepo:    outboxRepo,
	}
}

type TransformationResult struct {
	ID           uuid.UUID
	ImageID      uuid.UUID
	Spec         json.RawMessage
	Hash         string
	Status       transformDomain.Status
	ResultKey    string
	ErrorMessage string
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Duration     int64
	CreatedAt    time.Time
}

func (s *Service) RequestTransformation(ctx context.Context, imageID uuid.UUID, spec json.RawMessage) (*TransformationResult, error) {
	if _, err := s.imageRepo.GetByID(ctx, imageID); err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}

	t, err := transformDomain.NewTransformation(imageID, spec)
	if err != nil {
		return nil, err
	}

	existing, err := s.transformRepo.GetByImageAndHash(ctx, imageID, t.Hash())
	if err == nil {
		return toResult(existing), nil
	}
	if !errors.Is(err, transformDomain.ErrNotFound) {
		return nil, fmt.Errorf("check existing transformation: %w", err)
	}

	if err := s.txManager.WithTx(ctx, func(ctx context.Context, tx txtx.Tx) error {
		if err := s.transformRepo.Create(ctx, tx, t); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		if s.outboxRepo != nil {
			event := events.NewTransformationRequested(imageID, t.ID(), spec)
			payload, _ := json.Marshal(event)
			ev := &port.OutboxEvent{
				ID:          event.EventID,
				AggregateID: t.ID(),
				EventType:   events.EventTypeTransformationRequested,
				Payload:     payload,
				Status:      port.OutboxStatusPending,
				CreatedAt:   event.OccurredAt,
			}
			if err := s.outboxRepo.Save(ctx, tx, ev); err != nil {
				return fmt.Errorf("save outbox: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return toResult(t), nil
}

func (s *Service) GetTransformation(ctx context.Context, id uuid.UUID) (*TransformationResult, error) {
	t, err := s.transformRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResult(t), nil
}

func toResult(t *transformDomain.Transformation) *TransformationResult {
	return &TransformationResult{
		ID:           t.ID(),
		ImageID:      t.ImageID(),
		Spec:         t.Spec(),
		Hash:         t.Hash(),
		Status:       t.Status(),
		ResultKey:    t.ResultKey(),
		ErrorMessage: t.ErrorMessage(),
		StartedAt:    t.StartedAt(),
		CompletedAt:  t.CompletedAt(),
		Duration:     t.Duration(),
		CreatedAt:    t.CreatedAt(),
	}
}
