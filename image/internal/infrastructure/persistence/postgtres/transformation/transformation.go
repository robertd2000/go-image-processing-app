package transformationpg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/domain/transformation"
	"go.uber.org/zap"
)

type repo struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewTransformationRepo(pool *pgxpool.Pool, logger *zap.SugaredLogger) *repo {
	return &repo{pool: pool, logger: logger}
}

func (r *repo) Create(ctx context.Context, tx txtx.Tx, t *transformation.Transformation) error {
	query := `
		INSERT INTO transformations (id, image_id, transform_spec, transform_hash, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`
	return tx.Exec(ctx, query,
		t.ID(), t.ImageID(),
		[]byte(t.Spec()), t.Hash(),
		string(t.Status()), t.CreatedAt(),
	)
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*transformation.Transformation, error) {
	query := `
		SELECT id, image_id, transform_spec, transform_hash, status,
		       COALESCE(result_key, '') as result_key,
		       COALESCE(error_message, '') as error_message,
		       started_at, completed_at, COALESCE(duration, 0) as duration,
		       created_at
		FROM transformations WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var (
		tid, iid       uuid.UUID
		specBytes      []byte
		hash, status   string
		resultKey, em  string
		started, compl *time.Time
		dur            int64
		createdAt      time.Time
	)
	if err := row.Scan(&tid, &iid, &specBytes, &hash, &status,
		&resultKey, &em,
		&started, &compl, &dur,
		&createdAt); err != nil {
		return nil, fmt.Errorf("get transformation: %w", err)
	}

	return transformation.RestoreTransformation(
		tid, iid,
		json.RawMessage(specBytes), hash,
		transformation.Status(status),
		resultKey, em,
		started, compl,
		dur, createdAt,
	)
}

func (r *repo) GetByImageAndHash(ctx context.Context, imageID uuid.UUID, hash string) (*transformation.Transformation, error) {
	query := `
		SELECT id, image_id, transform_spec, transform_hash, status,
		       COALESCE(result_key, '') as result_key,
		       COALESCE(error_message, '') as error_message,
		       started_at, completed_at, COALESCE(duration, 0) as duration,
		       created_at
		FROM transformations WHERE image_id = $1 AND transform_hash = $2
	`
	row := r.pool.QueryRow(ctx, query, imageID, hash)

	var (
		tid, iid       uuid.UUID
		specBytes      []byte
		hashOut, stat  string
		resultKey, em  string
		started, compl *time.Time
		dur            int64
		createdAt      time.Time
	)
	if err := row.Scan(&tid, &iid, &specBytes, &hashOut, &stat,
		&resultKey, &em,
		&started, &compl, &dur,
		&createdAt); err != nil {
		return nil, fmt.Errorf("get transformation by hash: %w", err)
	}

	return transformation.RestoreTransformation(
		tid, iid,
		json.RawMessage(specBytes), hashOut,
		transformation.Status(stat),
		resultKey, em,
		started, compl,
		dur, createdAt,
	)
}
