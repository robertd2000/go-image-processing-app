package txmanagermem

import (
	"context"
	"sync"

	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
)

var _ txtx.Tx = (*FakeTx)(nil)

type FakeTx struct {
	mu         sync.Mutex
	committed  bool
	rolledBack bool
	pending    []func(ctx context.Context) error
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

func (t *FakeTx) OnCommit(fn func(ctx context.Context) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pending = append(t.pending, fn)
}

func (t *FakeTx) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fn := range t.pending {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	t.committed = true
	return nil
}

func (t *FakeTx) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rolledBack = true
	t.pending = nil
	return nil
}

func (t *FakeTx) Exec(ctx context.Context, query string, args ...any) error {
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
