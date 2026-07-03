package port

import (
	"context"

	txtx "github.com/robertd2000/go-image-processing-app/processor/internal/domain/tx"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx txtx.Tx) error) error
}
