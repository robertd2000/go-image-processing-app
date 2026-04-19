// Package txmanagerpg
package txmanagerpg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type TxWrapper struct {
	Tx pgx.Tx
}

func (t *TxWrapper) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *TxWrapper) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

func (t *TxWrapper) Exec(ctx context.Context, query string, args ...any) error {
	_, err := t.Tx.Exec(ctx, query, args...)
	return err
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx port.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	wrapped := &TxWrapper{Tx: tx}

	if err := fn(ctx, wrapped); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
