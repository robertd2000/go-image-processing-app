package request

import (
	"context"
	"errors"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
	"github.com/robertd2000/go-image-processing-app/processor/internal/usecase/request/model"
	"go.uber.org/zap"
)

type requestTransformationService struct {
	repo      transformation.Repository
	txManager port.TxManager
	logger    *zap.SugaredLogger
	metrics   port.Metrics
}

func NewRequestService(
	repo transformation.Repository,
	txManager port.TxManager,
	logger *zap.SugaredLogger,
	metrics port.Metrics,
) *requestTransformationService {
	return &requestTransformationService{
		repo:      repo,
		txManager: txManager,
		logger:    logger,
		metrics:   metrics,
	}
}

func (s *requestTransformationService) Execute(ctx context.Context, cmd model.Command) (*model.Result, error) {
	t, err := transformation.NewTransformation(cmd.ImageID, cmd.Source, cmd.Spec)
	if err != nil {
		return nil, fmt.Errorf("new transformation: %w", err)
	}

	existing, err := s.repo.GetByImageAndHash(ctx, t.ImageID(), t.Hash())
	switch {
	case err == nil:
		return model.ToResult(existing), nil

	case errors.Is(err, transformation.ErrNotFound):

	default:
		return nil, fmt.Errorf("get transformation by hash: %w", err)
	}

	if err := s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {
		err := s.repo.Create(ctx, tx, t)
		if err != nil {
			return fmt.Errorf("requestService: Create transformation: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return model.ToResult(t), nil
}
