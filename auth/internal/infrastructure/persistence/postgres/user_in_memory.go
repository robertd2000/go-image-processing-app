// Package postgres
package postgres

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
)

type userInMemoryRepository struct {
	mu   *sync.RWMutex
	data map[uuid.UUID]*userDomain.User
}

func NewUserInMemoryRepository() userDomain.UserRepository {
	return &userInMemoryRepository{
		data: make(map[uuid.UUID]*userDomain.User),
		mu:   &sync.RWMutex{},
	}
}

func (r *userInMemoryRepository) Create(_ context.Context, user *userDomain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existedUser, _ := r.findByEmail(context.Background(), user.Email())
	if existedUser != nil {
		return errors.New("user already exists")
	}
	r.data[user.ID()] = user
	return nil
}

func (r *userInMemoryRepository) Update(_ context.Context, user *userDomain.User) error {
	return nil
}

func (r *userInMemoryRepository) Delete(_ context.Context, id uuid.UUID) error {
	return nil
}

func (r *userInMemoryRepository) GetByUsername(_ context.Context, username string) (*userDomain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByUsername(context.Background(), username)
}

func (r *userInMemoryRepository) GetByEmail(_ context.Context, email string) (*userDomain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByEmail(context.Background(), email)
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByEmail(_ context.Context, email string) (*userDomain.User, error) {
	for _, user := range r.data {
		if user.Email() == email {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByUsername(_ context.Context, username string) (*userDomain.User, error) {
	for _, user := range r.data {
		if user.Username() == username {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// helper method to find user by email without locking
func (r *userInMemoryRepository) findByID(_ context.Context, userID uuid.UUID) (*userDomain.User, error) {
	user, exists := r.data[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *userInMemoryRepository) GetByID(ctx context.Context, userID uuid.UUID) (*userDomain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findByID(ctx, userID)
}

func (r *userInMemoryRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, err := r.findByEmail(context.Background(), email)

	return user != nil, err
}

func (r *userInMemoryRepository) ExistsByUsername(_ context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return false, nil
}
