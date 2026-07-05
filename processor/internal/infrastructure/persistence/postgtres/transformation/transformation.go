package transformation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	transformationDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
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

func (r *transformRepository) Create(ctx context.Context, tx port.Tx, t *transformationDomain.Transformation) error {
	spec, err := json.Marshal(t.Spec())
	if err != nil {
		return fmt.Errorf("marshal transform spec: %w", err)
	}

	var (
		resultStorageKey any
		resultMimeType   any
		resultWidth      any
		resultHeight     any
		resultSize       any
	)

	if result := t.Result(); result != nil {
		resultStorageKey = result.StorageKey()
		resultMimeType = result.MimeType()
		resultWidth = result.Width()
		resultHeight = result.Height()
		resultSize = result.Size()
	}

	var (
		errorMessage any
		startedAt    any
		completedAt  any
	)

	if t.ErrorMessage() != "" {
		errorMessage = t.ErrorMessage()
	}

	if t.StartedAt() != nil {
		startedAt = *t.StartedAt()
	}

	if t.CompletedAt() != nil {
		completedAt = *t.CompletedAt()
	}

	_, err = tx.Exec(
		ctx,
		insertTransformation,

		t.ID(),
		t.ImageID(),

		t.Source().StorageKey(),
		t.Source().MimeType(),
		t.Source().Width(),
		t.Source().Height(),

		spec,
		t.Hash(),

		t.Status(),

		resultStorageKey,
		resultMimeType,
		resultWidth,
		resultHeight,
		resultSize,

		errorMessage,

		startedAt,
		completedAt,

		t.CreatedAt(),
		t.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert transformation: %w", err)
	}

	return nil
}

func (r *transformRepository) GetByID(ctx context.Context, id uuid.UUID) (*transformationDomain.Transformation, error) {
	row := r.db.QueryRow(ctx, getTransformationByID, id)

	t, err := scanTransformation(row)
	if err != nil {
		return nil, fmt.Errorf("get transformation by id: %w", err)
	}

	return t, nil
}

func (r *transformRepository) Update(
	ctx context.Context,
	tx port.Tx,
	t *transformDomain.Transformation,
) error {

	var (
		resultStorageKey any
		resultMimeType   any
		resultWidth      any
		resultHeight     any
		resultSize       any
	)

	if result := t.Result(); result != nil {
		resultStorageKey = result.StorageKey()
		resultMimeType = result.MimeType()
		resultWidth = result.Width()
		resultHeight = result.Height()
		resultSize = result.Size()
	}

	var (
		errorMessage any
		startedAt    any
		completedAt  any
	)

	if t.ErrorMessage() != "" {
		errorMessage = t.ErrorMessage()
	}

	if t.StartedAt() != nil {
		startedAt = *t.StartedAt()
	}

	if t.CompletedAt() != nil {
		completedAt = *t.CompletedAt()
	}

	_, err := tx.Exec(
		ctx,
		updateTransformation,

		t.ID(),

		t.Status(),

		resultStorageKey,
		resultMimeType,
		resultWidth,
		resultHeight,
		resultSize,

		errorMessage,

		startedAt,
		completedAt,

		t.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update transformation: %w", err)
	}

	return nil
}
