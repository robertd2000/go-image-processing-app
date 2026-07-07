package transformation

import (
	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
	"go.uber.org/zap"
)

type transformationService struct {
	transformRepo transformDomain.Repository
	storage       port.Storage
	txManager     port.TxManager
	logger        *zap.SugaredLogger
	metrics       port.Metrics
}

func NewTransformationService(
	transformRepo transformDomain.Repository,
	storage port.Storage,
	txManager port.TxManager,
	logger *zap.SugaredLogger,
	metrics port.Metrics,
) *transformationService {
	return &transformationService{
		transformRepo: transformRepo,
		storage:       storage,
		txManager:     txManager,
		logger:        logger,
		metrics:       metrics,
	}
}
