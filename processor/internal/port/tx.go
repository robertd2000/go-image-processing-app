package port

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

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
