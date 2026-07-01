package tx

import "context"

type Row interface {
	Scan(dest ...any) error
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	Exec(ctx context.Context, query string, args ...any) error
	QueryRow(ctx context.Context, query string, args ...any) Row
}
