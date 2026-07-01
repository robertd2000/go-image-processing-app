package userrole

import (
	"context"

	"github.com/google/uuid"

	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

type Repository interface {
	Assign(
		ctx context.Context,
		tx txtx.Tx,
		userID uuid.UUID,
		roleID uuid.UUID,
	) error
}
