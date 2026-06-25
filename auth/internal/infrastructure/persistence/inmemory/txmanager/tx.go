package txmanagermem

import (
	"context"
	"sync"

	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

var _ txtx.Tx = (*FakeTx)(nil)

type FakeTx struct {
	mu         sync.Mutex
	committed  bool
	rolledBack bool
}

func (t *FakeTx) Committed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.committed
}

func (t *FakeTx) RolledBack() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rolledBack
}

func (t *FakeTx) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.committed = true
	return nil
}

func (t *FakeTx) Exec(ctx context.Context, query string, args ...any) error {
	return nil
}

func (t *FakeTx) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rolledBack = true
	return nil
}

type FakeTxManager struct {
	mu     sync.Mutex
	lastTx *FakeTx
}

func NewFakeTxManager() *FakeTxManager {
	return &FakeTxManager{}
}

func (m *FakeTxManager) LastTx() *FakeTx {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastTx
}

func (m *FakeTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, tx txtx.Tx) error) error {
	tx := &FakeTx{}

	m.mu.Lock()
	m.lastTx = tx
	m.mu.Unlock()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
