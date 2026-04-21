package kafkamiddleware

import (
	"context"
	"time"

	kafkahandler "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/kafka"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
)

type RetryConfig struct {
	MaxAttempts int
	Backoff     time.Duration
}

func RetryMiddleware(cfg RetryConfig) kafkahandler.Middleware {
	return func(next kafkahandler.Handler) kafkahandler.Handler {
		return HandlerFunc(func(ctx context.Context, evt events.Event) error {
			var err error

			for i := 1; i <= cfg.MaxAttempts; i++ {
				err = next.Handle(ctx, evt)
				if err == nil {
					return nil
				}

				time.Sleep(cfg.Backoff * time.Duration(i))
			}

			return err
		})
	}
}

func DLQMiddleware(dlq *DLQProducer) kafkahandler.Middleware {
	return func(next kafkahandler.Handler) kafkahandler.Handler {
		return HandlerFunc(func(ctx context.Context, evt events.Event) error {
			err := next.Handle(ctx, evt)
			if err != nil {
				_ = dlq.Send(ctx, evt, err)
				return nil
			}
			return nil
		})
	}
}

// helper
type HandlerFunc func(ctx context.Context, evt events.Event) error

func (f HandlerFunc) Handle(ctx context.Context, evt events.Event) error {
	return f(ctx, evt)
}
