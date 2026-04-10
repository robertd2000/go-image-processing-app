// Package usermem
package usermem

import (
	"context"
	"sync"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
)

type userInMemoryRepository struct {
	mu   *sync.RWMutex
	data map[uuid.UUID]*userDomain.AuthUser
}

func NewUserRepository() userDomain.UserRepository {
	return &userInMemoryRepository{
		data: make(map[uuid.UUID]*userDomain.AuthUser),
		mu:   &sync.RWMutex{},
	}
}

func (r *userInMemoryRepository) Create(_ context.Context, tx port.Tx, user *userDomain.AuthUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID() == uuid.Nil {
		return userDomain.ErrWrongCredentials
	}

	if user.Email() == nil {
		return userDomain.ErrWrongCredentials
	}

	existedUser, _ := r.findByEmail(context.Background(), *user.Email())
	if existedUser != nil {
		return userDomain.ErrUserAlreadyExists
	}

	r.data[user.ID()] = user
	return nil
}

func (r *userInMemoryRepository) Update(_ context.Context, user *userDomain.AuthUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID() == uuid.Nil {
		return userDomain.ErrWrongCredentials
	}

	if user.Email() == nil {
		return userDomain.ErrWrongCredentials
	}

	existedUser, _ := r.findByEmail(context.Background(), *user.Email())
	if existedUser != nil && existedUser.ID() != user.ID() {
		return userDomain.ErrUserAlreadyExists
	}

	r.data[user.ID()] = user
	return nil
}

func (r *userInMemoryRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)

	return nil
}

func (r *userInMemoryRepository) GetByUsername(_ context.Context, username string) (*userDomain.AuthUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByUsername(context.Background(), username)
}

func (r *userInMemoryRepository) GetByEmail(_ context.Context, email string) (*userDomain.AuthUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByEmail(context.Background(), email)
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByEmail(_ context.Context, email string) (*userDomain.AuthUser, error) {
	for _, user := range r.data {
		if *user.Email() == email {
			return user, nil
		}
	}

	return nil, userDomain.ErrUserNotFound
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByUsername(_ context.Context, username string) (*userDomain.AuthUser, error) {
	for _, user := range r.data {
		if user.Username() == username {
			return user, nil
		}
	}

	return nil, userDomain.ErrUserNotFound
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByID(_ context.Context, userID uuid.UUID) (*userDomain.AuthUser, error) {
	user, exists := r.data[userID]
	if !exists {
		return nil, userDomain.ErrUserNotFound
	}
	return user, nil
}

func (r *userInMemoryRepository) GetByID(ctx context.Context, userID uuid.UUID) (*userDomain.AuthUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByID(ctx, userID)
}

func (r *userInMemoryRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, _ := r.findByEmail(context.Background(), email)

	return user != nil, nil
}

func (r *userInMemoryRepository) ExistsByUsername(_ context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, _ := r.findByUsername(context.Background(), username)

	return user != nil, nil
}
