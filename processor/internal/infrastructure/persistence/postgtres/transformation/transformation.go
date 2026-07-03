package transformation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/port"
)

const transformationColumns = `
	id,
	image_id,
	storage_key,
	mime_type,
	width,
	height,
	transform_spec,
	transform_hash,
	status,
	result_key,
	error_message,
	started_at,
	completed_at,
	created_at
`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	tx txtx.Tx,
	t *transformDomain.Transformation,
) error {
	const query = `
		INSERT INTO transformations (
			id,
			image_id,
			storage_key,
			mime_type,
			width,
			height,
			transform_spec,
			transform_hash,
			status,
			result_key,
			error_message,
			started_at,
			completed_at,
			created_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,$11,$12,$13,$14
		)
	`

	_, err := tx.Exec(
		ctx,
		query,
		t.ID(),
		t.ImageID(),
		t.StorageKey(),
		t.MimeType(),
		t.Width(),
		t.Height(),
		t.Spec(),
		t.Hash(),
		t.Status(),
		t.ResultKey(),
		t.ErrorMessage(),
		t.StartedAt(),
		t.CompletedAt(),
		t.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("create transformation: %w", err)
	}

	return nil
}

func (r *Repository) Update(
	ctx context.Context,
	tx txtx.Tx,
	t *transformDomain.Transformation,
) error {
	const query = `
		UPDATE transformations
		SET
			status = $2,
			result_key = $3,
			error_message = $4,
			started_at = $5,
			completed_at = $6
		WHERE id = $1
	`

	tag, err := tx.Exec(
		ctx,
		query,
		t.ID(),
		t.Status(),
		t.ResultKey(),
		t.ErrorMessage(),
		t.StartedAt(),
		t.CompletedAt(),
	)
	if err != nil {
		return fmt.Errorf("update transformation: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return transformDomain.ErrNotFound
	}

	return nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*transformDomain.Transformation, error) {

	query := `
		SELECT ` + transformationColumns + `
		FROM transformations
		WHERE id = $1
	`

	return scanTransformation(
		r.db.QueryRow(ctx, query, id),
	)
}

func (r *Repository) GetByImageAndHash(
	ctx context.Context,
	imageID uuid.UUID,
	hash string,
) (*transformDomain.Transformation, error) {

	query := `
		SELECT ` + transformationColumns + `
		FROM transformations
		WHERE image_id = $1
		  AND transform_hash = $2
	`

	return scanTransformation(
		r.db.QueryRow(ctx, query, imageID, hash),
	)
}

func (r *Repository) GetPending(
	ctx context.Context,
	limit int,
) ([]*transformDomain.Transformation, error) {

	query := `
		SELECT ` + transformationColumns + `
		FROM transformations
		WHERE status = $1
		ORDER BY created_at
		LIMIT $2
	`

	rows, err := r.db.Query(
		ctx,
		query,
		transformDomain.StatusPending,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query pending transformations: %w", err)
	}
	defer rows.Close()

	result := make([]*transformDomain.Transformation, 0)

	for rows.Next() {
		t, err := scanTransformation(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return result, nil
}

func scanTransformation(row interface {
	Scan(dest ...any) error
}) (*transformDomain.Transformation, error) {

	var (
		id uuid.UUID

		imageID uuid.UUID

		storageKey string
		mimeType   string

		width  int
		height int

		spec []byte

		hash string

		status transformDomain.Status

		resultKey string
		errorMsg  string

		startedAt   *time.Time
		completedAt *time.Time

		createdAt time.Time
	)

	err := row.Scan(
		&id,
		&imageID,
		&storageKey,
		&mimeType,
		&width,
		&height,
		&spec,
		&hash,
		&status,
		&resultKey,
		&errorMsg,
		&startedAt,
		&completedAt,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, transformDomain.ErrNotFound
		}

		return nil, fmt.Errorf("scan transformation: %w", err)
	}

	return transformDomain.RestoreTransformation(
		id,
		imageID,
		storageKey,
		mimeType,
		width,
		height,
		json.RawMessage(spec),
		hash,
		status,
		resultKey,
		errorMsg,
		startedAt,
		completedAt,
		createdAt,
	)
}
