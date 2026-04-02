// Package usermem
package usermem

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	userDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/user"
)

type userInMemoryRepository struct {
	users map[uuid.UUID]*userDomain.User
	mu    *sync.RWMutex
}

// Count implements user.UserRepository.
func (u *userInMemoryRepository) Count(ctx context.Context, filter userDomain.UserFilter) (int, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	count := 0
	for _, user := range u.users {
		if u.matchesFilter(user, filter) {
			count++
		}
	}

	return count, nil
}

// List implements user.UserRepository.
func (u *userInMemoryRepository) List(ctx context.Context, filter userDomain.UserFilter) ([]*userDomain.User, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	filtered := make([]*userDomain.User, 0, len(u.users))
	for _, user := range u.users {
		if u.matchesFilter(user, filter) {
			filtered = append(filtered, user)
		}
	}

	if filter.SortBy != "" {
		order := strings.ToLower(filter.SortOrder)
		if order != "desc" {
			order = "asc"
		}

		sort.SliceStable(filtered, func(i, j int) bool {
			var lhs, rhs string
			switch strings.ToLower(filter.SortBy) {
			case "username":
				lhs = filtered[i].Username().String()
				rhs = filtered[j].Username().String()
			case "email":
				lhs = filtered[i].Email().String()
				rhs = filtered[j].Email().String()
			default:
				return true
			}

			if order == "asc" {
				return strings.ToLower(lhs) < strings.ToLower(rhs)
			}
			return strings.ToLower(lhs) > strings.ToLower(rhs)
		})
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	limit := filter.Limit
	if limit < 0 {
		limit = 0
	}

	if offset >= len(filtered) {
		return []*userDomain.User{}, nil
	}

	end := len(filtered)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	return filtered[offset:end], nil
}

func (u *userInMemoryRepository) matchesFilter(user *userDomain.User, filter userDomain.UserFilter) bool {
	if filter.Status != nil && user.Status() != *filter.Status {
		return false
	}

	if filter.Search != nil && *filter.Search != "" {
		search := strings.ToLower(*filter.Search)
		username := strings.ToLower(user.Username().String())
		email := strings.ToLower(user.Email().String())
		if !strings.Contains(username, search) && !strings.Contains(email, search) {
			return false
		}
	}

	return true
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
