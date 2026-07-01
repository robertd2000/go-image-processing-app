package userrolemem

import (
	"context"

	"github.com/google/uuid"
	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

type Assignment struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

type FakeRepository struct {
	assignments []Assignment
}

func (r *FakeRepository) Assign(
	ctx context.Context,
	tx txtx.Tx,
	userID uuid.UUID,
	roleID uuid.UUID,
) error {

	r.assignments = append(r.assignments, Assignment{
		UserID: userID,
		RoleID: roleID,
	})

	return nil
}
