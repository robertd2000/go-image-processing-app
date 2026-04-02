// Package usermem
package usermem

import (
	"context"
	"sync"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
)

type userInMemoryRepository struct {
	users map[uuid.UUID]*userDomain.User
	mu    *sync.RWMutex
}

// FindByUsername implements user.UserRepository.
func (u *userInMemoryRepository) FindByUsername(ctx context.Context, username userDomain.Username) (*userDomain.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, user := range u.users {
		if user.Username() == username {
			if user.Status() == userDomain.StatusInactive {
				return nil, userDomain.ErrUserNotFound
			}
			return user, nil
		}
	}
	return nil, userDomain.ErrUserNotFound
}

// ExistsByEmail implements user.UserRepository.ExistsByEmail
func (u *userInMemoryRepository) ExistsByEmail(ctx context.Context, email userDomain.Email) (bool, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, user := range u.users {
		if user.Email().String() == email.String() {
			return true, nil
		}
	}
	return false, nil
}

// ExistsByUsername implements user.UserRepository.
func (u *userInMemoryRepository) ExistsByUsername(ctx context.Context, username userDomain.Username) (bool, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, user := range u.users {
		if user.Username().String() == username.String() {
			return true, nil
		}
	}
	return false, nil
}

// Create implements user.UserRepository.
func (u *userInMemoryRepository) Create(ctx context.Context, user *userDomain.User) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.users[user.ID()] = user

	return nil
}

// Delete implements user.UserRepository.
func (u *userInMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	user, err := u.findByID(ctx, id)
	if err != nil {
		return err
	}

	user.Deactivate()

	u.users[id] = user

	return nil
}

// FindByEmail implements user.UserRepository.
func (u *userInMemoryRepository) FindByEmail(ctx context.Context, email userDomain.Email) (*userDomain.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	for _, user := range u.users {
		if user.Email() == email {
			if user.Status() == userDomain.StatusInactive {
				return nil, userDomain.ErrUserNotFound
			}
			return user, nil
		}
	}
	return nil, userDomain.ErrUserNotFound
}

// FindByID implements user.UserRepository.
func (u *userInMemoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.findByID(ctx, id)
}

// Update implements user.UserRepository.
func (u *userInMemoryRepository) Update(ctx context.Context, user *userDomain.User) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, exists := u.users[user.ID()]; !exists {
		return userDomain.ErrUserNotFound
	}

	u.users[user.ID()] = user
	return nil
}

func (u *userInMemoryRepository) findByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	if user, exists := u.users[id]; exists {
		return user, nil
	}
	return nil, userDomain.ErrUserNotFound
}

func NewUserRepository() userDomain.UserRepository {
	return &userInMemoryRepository{
		users: make(map[uuid.UUID]*userDomain.User),
		mu:    &sync.RWMutex{},
	}
}
