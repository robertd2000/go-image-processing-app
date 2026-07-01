package rolemem

import (
	"context"
	"sync"

	"github.com/google/uuid"

	roleDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/role"
	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
)

var _ roleDomain.Repository = (*FakeRepository)(nil)

type FakeRepository struct {
	mu    sync.RWMutex
	roles map[uuid.UUID]*roleDomain.Role
}

func NewRoleRepository() *FakeRepository {
	r := &FakeRepository{
		roles: make(map[uuid.UUID]*roleDomain.Role),
	}

	admin, err := roleDomain.FromName(uuid.New(), roleDomain.Admin)
	if err != nil {
		panic(err)
	}

	user, err := roleDomain.FromName(uuid.New(), roleDomain.User)
	if err != nil {
		panic(err)
	}

	r.roles[admin.ID()] = admin
	r.roles[user.ID()] = user

	return r
}

func (r *FakeRepository) Save(role *roleDomain.Role) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.roles[role.ID()] = role
}

func (r *FakeRepository) ByID(
	ctx context.Context,
	tx txtx.Tx,
	id uuid.UUID,
) (*roleDomain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, ok := r.roles[id]
	if !ok {
		return nil, roleDomain.ErrRoleNotFound
	}

	return role, nil
}

func (r *FakeRepository) ByName(
	ctx context.Context,
	tx txtx.Tx,
	name roleDomain.Name,
) (*roleDomain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, role := range r.roles {
		if role.Name() == name {
			return role, nil
		}
	}

	return nil, roleDomain.ErrRoleNotFound
}
