package transformation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
)

func scanTransformation(row pgx.Row) (*transformDomain.Transformation, error) {
	var (
		id      uuid.UUID
		imageID uuid.UUID

		sourceStorageKey string
		sourceMimeType   string
		sourceWidth      int
		sourceHeight     int

		specData []byte
		hash     string
		status   string

		resultStorageKey sql.NullString
		resultMimeType   sql.NullString
		resultWidth      sql.NullInt32
		resultHeight     sql.NullInt32
		resultSize       sql.NullInt64

		errorMessage sql.NullString

		startedAt   sql.NullTime
		completedAt sql.NullTime

		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(
		&id,
		&imageID,

		&sourceStorageKey,
		&sourceMimeType,
		&sourceWidth,
		&sourceHeight,

		&specData,
		&hash,

		&status,

		&resultStorageKey,
		&resultMimeType,
		&resultWidth,
		&resultHeight,
		&resultSize,

		&errorMessage,

		&startedAt,
		&completedAt,

		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, transformDomain.ErrNotFound
		}
		return nil, err
	}

	var spec transformDomain.TransformSpec
	if err := json.Unmarshal(specData, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal transform spec: %w", err)
	}

	source, err := transformDomain.NewSourceImage(
		sourceStorageKey,
		sourceMimeType,
		sourceWidth,
		sourceHeight,
	)
	if err != nil {
		return nil, fmt.Errorf("restore source image: %w", err)
	}

	var result *transformDomain.ResultImage

	if resultStorageKey.Valid {
		r, err := transformDomain.NewResultImage(
			resultStorageKey.String,
			resultMimeType.String,
			int(resultWidth.Int32),
			int(resultHeight.Int32),
			resultSize.Int64,
		)
		if err != nil {
			return nil, fmt.Errorf("restore result image: %w", err)
		}

		result = &r
	}

	var errMsg string
	if errorMessage.Valid {
		errMsg = errorMessage.String
	}

	var started *time.Time
	if startedAt.Valid {
		t := startedAt.Time
		started = &t
	}

	var completed *time.Time
	if completedAt.Valid {
		t := completedAt.Time
		completed = &t
	}

	return transformDomain.RestoreTransformation(
		id,
		imageID,
		source,
		spec,
		hash,
		transformDomain.Status(status),
		result,
		errMsg,
		started,
		completed,
		createdAt,
		updatedAt,
	)
}
