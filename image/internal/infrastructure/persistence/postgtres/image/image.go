// Package imagepg
package imagepg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
)

type imageRepository struct {
	db *pgxpool.Pool
}

func NewImageRepository(db *pgxpool.Pool) *imageRepository {
	return &imageRepository{
		db: db,
	}
}

func (r *imageRepository) Save(ctx context.Context, image *imageDomain.Image) error {
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

	_, err := r.db.Exec(
		ctx,
		query,
		image.ID(),
		image.UserID(),
		image.OriginalName(),
		image.StorageKey(),
		meta.Size(),
		meta.MimeType(),
		meta.Width(),
		meta.Height(),
		image.CreatedAt(),
	)
	if err != nil {
		return mapPGError(fmt.Errorf("image repository: save: %w", err))
	}

	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID) (*imageDomain.Image, error) {
	return nil, nil
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
