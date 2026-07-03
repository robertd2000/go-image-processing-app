package port

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type DB interface {
	Exec(
		ctx context.Context,
		query string,
		args ...any,
	) (pgconn.CommandTag, error)

	Query(
		ctx context.Context,
		query string,
		args ...any,
	) (pgx.Rows, error)

	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) pgx.Row
}
