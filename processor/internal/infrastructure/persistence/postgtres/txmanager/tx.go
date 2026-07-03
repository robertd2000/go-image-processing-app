// Package txmanagerpg
package txmanagerpg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/robertd2000/go-image-processing-app/processor/internal/port"

	"go.uber.org/zap"
)

var (
	_ port.Tx        = (*TxWrapper)(nil)
	_ port.TxManager = (*TxManager)(nil)
)

type TxWrapper struct {
	tx pgx.Tx
}

func NewTxWrapper(tx pgx.Tx) *TxWrapper {
	return &TxWrapper{
		tx: tx,
	}
}

func (t *TxWrapper) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *TxWrapper) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *TxWrapper) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, query, args...)
}

func (t *TxWrapper) Query(
	ctx context.Context,
	query string,
	args ...any,
) (pgx.Rows, error) {
	return t.tx.Query(ctx, query, args...)
}

func (t *TxWrapper) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) pgx.Row {
	return t.tx.QueryRow(ctx, query, args...)
}

type TxManager struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewTxManager(
	pool *pgxpool.Pool,
	logger *zap.SugaredLogger,
) *TxManager {
	return &TxManager{
		pool:   pool,
		logger: logger,
	}
}

func (m *TxManager) WithTx(
	ctx context.Context,
	fn func(ctx context.Context, tx port.Tx) error,
) error {

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.logger.Errorw("failed to begin transaction", "error", err)
		return fmt.Errorf("begin transaction: %w", err)
	}

	wrapped := NewTxWrapper(tx)

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(ctx, wrapped); err != nil {

		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			m.logger.Errorw(
				"rollback failed",
				"rollback_error", rollbackErr,
				"original_error", err,
			)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		m.logger.Errorw("commit failed", "error", err)
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
