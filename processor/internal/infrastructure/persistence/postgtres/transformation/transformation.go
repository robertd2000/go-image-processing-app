package transformation

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/processor/internal/port"
	"go.uber.org/zap"
)

type transformRepository struct {
	db      *pgxpool.Pool
	logger  *zap.SugaredLogger
	metrics port.Metrics
}

func NewTransformRepository(db *pgxpool.Pool, logger *zap.SugaredLogger, metrics port.Metrics) *transformRepository {
	return &transformRepository{
		db:      db,
		logger:  logger,
		metrics: metrics,
	}
}
