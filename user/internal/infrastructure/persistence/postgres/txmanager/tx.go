// Package txmanagerpg
package txmanagerpg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/user/internal/port"
	"go.uber.org/zap"
)

type TxWrapper struct {
	Tx pgx.Tx
}

func NewTxWrapper(tx pgx.Tx) *TxWrapper {
	return &TxWrapper{Tx: tx}
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
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewTxManager(pool *pgxpool.Pool, logger *zap.SugaredLogger) *TxManager {
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
		m.logger.Errorw("failed to begin tx", "error", err)
		return fmt.Errorf("begin tx: %w", err)
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
			m.logger.Errorw("rollback failed",
				"error", rollbackErr,
				"original_error", err,
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		m.logger.Errorw("commit failed", "error", err)
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
