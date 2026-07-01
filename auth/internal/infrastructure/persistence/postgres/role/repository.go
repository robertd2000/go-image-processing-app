package rolepg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	roleDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

type Repository struct{}

func NewRoleRepository() *Repository {
	return &Repository{}
}

func (r *Repository) ByID(
	ctx context.Context,
	tx txtx.Tx,
	id uuid.UUID,
) (*roleDomain.Role, error) {

	const query = `
		SELECT id, name
		FROM roles
		WHERE id = $1
	`

	var (
		roleID uuid.UUID
		name   string
	)

	err := tx.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&roleID,
		&name,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, roleDomain.ErrRoleNotFound
		}
		return nil, err
	}

	return roleDomain.FromName(
		roleID,
		roleDomain.Name(name),
	)
}

func (r *Repository) ByName(
	ctx context.Context,
	tx txtx.Tx,
	name roleDomain.Name,
) (*roleDomain.Role, error) {

	const query = `
		SELECT id, name
		FROM roles
		WHERE name = $1
	`

	var (
		roleID   uuid.UUID
		roleName string
	)

	err := tx.QueryRow(
		ctx,
		query,
		name,
	).Scan(
		&roleID,
		&roleName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, roleDomain.ErrRoleNotFound
		}
		return nil, err
	}

	return roleDomain.FromName(
		roleID,
		roleDomain.Name(roleName),
	)
}
