package app

import (
	"context"
	"errors"
	"sync"
)

type Closer interface {
	Close(ctx context.Context) error
}

type CloserFunc func(ctx context.Context) error

func (f CloserFunc) Close(ctx context.Context) error { return f(ctx) }

type Lifecycle struct {
	mu      sync.Mutex
	closers []Closer
	wg      sync.WaitGroup
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

func (l *Lifecycle) Add(c Closer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closers = append(l.closers, c)
}

func (l *Lifecycle) Go(fn func(ctx context.Context), ctx context.Context) {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		fn(ctx)
	}()
}

func (l *Lifecycle) Shutdown(ctx context.Context) error {
	var errs []error

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
