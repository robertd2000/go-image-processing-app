// Package imagepg
package imagepg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"go.uber.org/zap"
)

type imageRepository struct {
	db      *pgxpool.Pool
	logger  *zap.SugaredLogger
	metrics port.Metrics
}

func NewImageRepository(db *pgxpool.Pool, logger *zap.SugaredLogger, metrics port.Metrics) *imageRepository {
	return &imageRepository{
		db:      db,
		logger:  logger,
		metrics: metrics,
	}
}

func (r *imageRepository) Save(ctx context.Context, tx port.Tx, image *imageDomain.Image) error {
	if image == nil {
		return fmt.Errorf("image repository: save: nil image")
	}

	meta := image.Metadata()

	query := `
		INSERT INTO images (
			id, user_id,
			original_name, storage_key,
			file_size, mime_type,
			width, height,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id
	`

	err := tx.Exec(
		ctx,
		query,
		image.ID(),
		image.UserID(),
		image.OriginalName(),
		string(image.StorageKey()),
		meta.Size(),
		meta.MimeType(),
		meta.Width(),
		meta.Height(),
		image.CreatedAt(),
	)
	if err != nil {
		r.metrics.IncImageSaveError()
		r.logger.Errorw("failed to save image",
			"image_id", image.ID(),
			"user_id", image.UserID(),
			"error", err,
		)

		return mapPGError(fmt.Errorf("image repository: save: %w", err))
	}

	r.metrics.IncImageSaveSuccess()

	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID) (*imageDomain.Image, error) {
	query := `
		SELECT 
			id, user_id,
			original_name, storage_key,
			file_size, mime_type,
			width, height,
			created_at
		FROM images
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var (
		imgID        uuid.UUID
		userID       uuid.UUID
		storageKey   string
		originalName string
		size         int64
		mimeType     string
		width        int
		height       int
		createdAt    time.Time
		deletedAt    *time.Time
	)

	err := row.Scan(
		&imgID,
		&userID,
		&storageKey,
		&originalName,
		&size,
		&mimeType,
		&width,
		&height,
		&createdAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, imageDomain.ErrNotFound
		}
		r.logger.Errorw("GetByID failed", "id", id, "error", err)
		return nil, mapPGError(err)
	}

	if deletedAt != nil {
		return nil, imageDomain.ErrNotFound
	}

	meta, err := imageDomain.NewImageMetadata(width, height, size, mimeType)
	if err != nil {
		return nil, err
	}

	img, err := imageDomain.RestoreImage(
		imgID,
		userID,
		imageDomain.StorageKey(storageKey),
		originalName,
		meta,
		createdAt,
		time.Time{},
	)

	return img, nil
}

func (r *imageRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	// start := time.Now()
	// defer r.metrics.ObserveDB("image.count_by_user", time.Since(start))

	query := `
		SELECT COUNT(*)
		FROM images
		WHERE user_id = $1
		AND deleted_at IS NULL
	`

	var count int

	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		r.logger.Errorw("CountByUser failed", "user_id", userID, "error", err)
		return 0, mapPGError(err)
	}

	return count, nil
}

func (r *imageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE images
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Errorw("Delete failed", "id", id, "error", err)
		return mapPGError(err)
	}

	if cmd.RowsAffected() == 0 {
		return imageDomain.ErrNotFound
	}

	return nil
}

func (r *imageRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*imageDomain.Image, error) {
	return []*imageDomain.Image{}, nil
}

func mapPGError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {

	// unique_violation
	case "23505":
		return imageDomain.ErrAlreadyExists

	// foreign_key_violation
	case "23503":
		return fmt.Errorf("image repository: invalid reference: %w", err)
	}

	return fmt.Errorf("image repository: internal error: %w", err)
}
