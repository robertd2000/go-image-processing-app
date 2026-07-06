package request

import (
	"context"
	"fmt"

	"github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
	"github.com/robertd2000/go-image-processing-app/processor/internal/usecase/request/model"
	"go.uber.org/zap"
)

type requestService struct {
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
) *requestService {
	return &requestService{
		repo:      repo,
		txManager: txManager,
		logger:    logger,
		metrics:   metrics,
	}
}

func (s *requestService) Execute(ctx context.Context, cmd model.Command) (*model.Result, error) {
	transformation, err := transformation.NewTransformation(cmd.ImageID, cmd.Source, cmd.Spec)
	if err != nil {
		return nil, fmt.Errorf("requestService: get transformation: %w", err)
	}

	existedTransformation, err := s.repo.GetByImageAndHash(ctx, transformation.ImageID(), transformation.Hash())
	if err != nil {
		return nil, fmt.Errorf("requestService: %w", err)
	}
	if existedTransformation != nil {
		return model.ToResult(existedTransformation), nil
	}

	if err := s.txManager.WithTx(ctx, func(ctx context.Context, tx port.Tx) error {
		err := s.repo.Create(ctx, tx, transformation)
		if err != nil {
			return fmt.Errorf("requestService: Create transformation: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return model.ToResult(existedTransformation), nil
}
