package port

import (
	"context"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Exec(ctx context.Context, query string, args ...any) error
}
