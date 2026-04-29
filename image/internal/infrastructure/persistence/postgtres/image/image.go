// Package imagepg
package imagepg

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type imageRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *imageRepository {
	return &imageRepository{
		db: db,
	}
}
