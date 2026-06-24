package app

import (
	"context"
	"errors"
	"sync"
)

// Closer is a resource that needs cleanup on shutdown.
type Closer interface {
	Close(ctx context.Context) error
}

// CloserFunc adapts a function to Closer.
type CloserFunc func(ctx context.Context) error

func (f CloserFunc) Close(ctx context.Context) error { return f(ctx) }

// Lifecycle manages background goroutines and resource cleanup.
// Registration order defines shutdown order in reverse:
// first registered = last closed.
//
// Usage:
//
//	lc := app.NewLifecycle()
//	lc.Add(closer1)         // closed last
//	lc.Add(closer2)         // closed first
//	lc.Go(fn, appCtx)       // goroutine, stopped via cancel(appCtx)
//	cancel(appCtx)          // signal goroutines
//	lc.Shutdown(shutdownCtx) // wait goroutines → close resources
type Lifecycle struct {
	mu      sync.Mutex
	closers []Closer
	wg      sync.WaitGroup
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

// Add registers a closer for cleanup on Shutdown.
func (l *Lifecycle) Add(c Closer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closers = append(l.closers, c)
}

// Go launches fn in a tracked goroutine. fn must exit when ctx is cancelled.
func (l *Lifecycle) Go(fn func(ctx context.Context), ctx context.Context) {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		fn(ctx)
	}()
}

// Shutdown waits for all goroutines to finish, then closes resources
// in reverse registration order. Respects ctx cancellation and timeout.
func (l *Lifecycle) Shutdown(ctx context.Context) error {
	var errs []error

	// 1. Wait for goroutines
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		errs = append(errs, ctx.Err())
	}

	// 2. Close resources in reverse order
	l.mu.Lock()
	closers := make([]Closer, len(l.closers))
	copy(closers, l.closers)
	l.mu.Unlock()

	for i := len(closers) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return errors.Join(errs...)
		default:
		}
		if err := closers[i].Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
