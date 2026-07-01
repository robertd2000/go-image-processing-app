package auth_test

// Transactional test ensures rollback on failure during user registration.
// It substitutes a failing implementation of Create() for the in‑memory UserRepo.
// The test verifies that after an error, no user exists and outbox remains empty.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	txtx "github.com/robertd2000/go-image-processing-app/auth/internal/domain/tx"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	jwt "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	outboxmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/outbox"
	rolemem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/role"
	tokenmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
	txmanagermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/txmanager"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	userrolemem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/userrole"
	security "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	authsvc "github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"

	"github.com/stretchr/testify/assert"
)

// failingUserRepo wraps an existing UserRepository and forces Create to fail.
// The rest of the methods delegate to the underlying repo.

type failingUserRepo struct {
	userDomain.UserRepository
	fail bool
}

func (r *failingUserRepo) Create(ctx context.Context, tx txtx.Tx, u *userDomain.AuthUser) error { //nolint:revive
	if r.fail {
		return errors.New("forced failure")
	}
	return r.UserRepository.Create(ctx, tx, u)
}

func TestAuthService_TransactionRollbackOnFailure(t *testing.T) {
	ctx := context.Background()

	baseRepo := usermem.NewUserRepository()
	repo := &failingUserRepo{UserRepository: baseRepo, fail: true}

	tokenRepo := tokenmem.NewTokenRepository()
	outboxRepo := outboxmem.NewRepository()
	passwordHasher := &security.FakeHasher{}
	tokenHasher := &security.FakeTokenHasher{}
	txManager := txmanagermem.NewFakeTxManager()
	roleRepo := rolemem.NewRoleRepository()
	userRoleRepo := userrolemem.NewUserRoleRepository()

	svc := authsvc.NewAuthService(
		repo,
		tokenRepo,
		roleRepo,
		userRoleRepo,
		outboxRepo,
		passwordHasher,
		tokenHasher,
		jwt.NewInMemoryTokenGenerator(),
		10*time.Minute, // access TTL - irrelevant for this test
		60*time.Minute, // refresh TTL
		txManager,
	)

	err := svc.Register(ctx, model.RegisterInput{
		Username: "faileduser",
		Email:    "fd@example.com",
		Password: "Password123!",
	})
	assert.Error(t, err)

	// User should not be persisted.
	_, err = baseRepo.GetByEmail(ctx, "fd@example.com")
	assert.Error(t, err)

	// Outbox should remain empty.
	evs, _ := outboxRepo.GetUnprocessed(ctx, 10)
	assert.Empty(t, evs)
}

// TestAuthService_OutboxEventCreation validates that a registration emits the correct event.
func TestAuthService_OutboxEventCreation(t *testing.T) {
	ctx := context.Background()
	userRepo := usermem.NewUserRepository()
	tokenRepo := tokenmem.NewTokenRepository()
	outboxRepo := outboxmem.NewRepository()
	passwordHasher := &security.FakeHasher{}
	tokenHasher := &security.FakeTokenHasher{}
	txManager := txmanagermem.NewFakeTxManager()

	roleRepo := rolemem.NewRoleRepository()
	userRoleRepo := userrolemem.NewUserRoleRepository()

	svc := authsvc.NewAuthService(
		userRepo,
		tokenRepo,
		roleRepo,
		userRoleRepo,
		outboxRepo,
		passwordHasher,
		tokenHasher,
		jwt.NewInMemoryTokenGenerator(),
		10*time.Minute,
		60*time.Minute,
		txManager,
	)

	err := svc.Register(ctx, model.RegisterInput{
		Username: "eventuser",
		Email:    "evt@example.com",
		Password: "Password123!",
	})
	assert.NoError(t, err)

	evs, _ := outboxRepo.GetUnprocessed(ctx, 1)
	assert.Len(t, evs, 1)

	evt := evs[0]
	const expectedType = "user.created"
	if evt.Type != expectedType {
		t.Fatalf("event type mismatch: got %s want %s", evt.Type, expectedType)
	}

	// Decode payload to map for basic assertions.
	var data map[string]interface{}
	err = json.Unmarshal(evt.Payload, &data)
	assert.NoError(t, err)
	// Verify that the event key matches the user ID.
	createdUser, err := userRepo.GetByEmail(ctx, "evt@example.com")
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID().String(), evt.Key)

}

// TestAuthService_SessionLimit enforces the per‑user session cap.
// func TestAuthService_SessionLimit(t *testing.T) { /* placeholder for future session limit test */ }
