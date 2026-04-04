package txmanagermem

import (
	"context"

	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type FakeTx struct{}

func (t *FakeTx) Commit(ctx context.Context) error                          { return nil }
func (t *FakeTx) Exec(ctx context.Context, query string, args ...any) error { return nil }
func (t *FakeTx) Rollback(ctx context.Context) error                        { return nil }

type FakeTxManager struct{}

func (m *FakeTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx port.Tx) error) error {
	return fn(ctx, &FakeTx{})
}
