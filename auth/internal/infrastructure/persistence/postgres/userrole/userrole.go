package userrole

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
	userroleDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/userrole"
	"go.uber.org/zap"
)

type userroleRepository struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewUserRepository(db *pgxpool.Pool, logger *zap.SugaredLogger) userroleDomain.Repository {
	return &userroleRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userroleRepository) Assign(
	ctx context.Context,
	tx txtx.Tx,
	userID uuid.UUID,
	roleID uuid.UUID,
) error {
	err := tx.Exec(
		ctx,
		`
        INSERT INTO user_roles (user_id, role_id)
        VALUES ($1, $2)
        `,
		userID,
		roleID,
	)

	return err
}
