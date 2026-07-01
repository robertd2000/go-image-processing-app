package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	tokensDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	userDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/user"
	jwt "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	outboxmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/outbox"
	tokenmem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/token"
	txmanagermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/txmanager"
	usermem "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/inmemory/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/port"
	authsvc "github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
	"github.com/robertd2000/go-image-processing-app/auth/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/google/uuid"
)

// MockPublisher implements port.EventPublisher for testing
type MockPublisher struct {
	PublishedEvents []PublishedEvent
	MockErr         error
}

type PublishedEvent struct {
	Topic string
	Key   []byte
	Msg   any
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key []byte, msg any) error {
	if m.MockErr != nil {
		return m.MockErr
	}
	m.PublishedEvents = append(m.PublishedEvents, PublishedEvent{
		Topic: topic,
		Key:   key,
		Msg:   msg,
	})
	return nil
}

func (m *MockPublisher) Reset() {
	m.PublishedEvents = nil
	m.MockErr = nil
}

type KafkaOutboxTestSuite struct {
	suite.Suite
	ctx       context.Context
	userRepo  userDomain.UserRepository
	tokenRepo tokensDomain.TokenRepository
	outbox    port.OutboxRepository
	txManager port.TxManager
	service   AuthService
}

func (s *KafkaOutboxTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepo = usermem.NewUserRepository()
	s.tokenRepo = tokenmem.NewTokenRepository()
	s.outbox = outboxmem.NewRepository()
	s.txManager = txmanagermem.NewFakeTxManager()
	s.service = authsvc.NewAuthService(
		s.userRepo,
		s.tokenRepo,
		s.outbox,
		&security.FakeHasher{},
		&security.FakeTokenHasher{},
		jwt.NewInMemoryTokenGenerator(),
		10*time.Minute,
		60*time.Minute,
		s.txManager,
	)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventTypeIsUserCreated() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_user1",
		Email:    "kafka1@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, err := s.outbox.GetUnprocessed(s.ctx, 10)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), evts, 1)
	assert.Equal(s.T(), "user.created", evts[0].Type)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventTopicIsUserEvents() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_user2",
		Email:    "kafka2@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Equal(events.UserEventsTopic, evts[0].Topic)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventKeyMatchesUserID() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_user3",
		Email:    "kafka3@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Len(evts, 1)

	user, err := s.userRepo.GetByEmail(s.ctx, "kafka3@example.com")
	s.Require().NoError(err)

	s.Equal(user.ID().String(), evts[0].Key)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventPayloadDecodesCorrectly() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username:  "kafka_user4",
		Email:     "kafka4@example.com",
		Password:  "Secure123!",
		FirstName: "John",
		LastName:  "Doe",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Require().Len(evts, 1)

	var rawEvent events.Event
	err = json.Unmarshal(evts[0].Payload, &rawEvent)
	assert.NoError(s.T(), err)

	assert.NotEmpty(s.T(), rawEvent.EventID)
	assert.Equal(s.T(), "user.created", rawEvent.EventType)
	assert.Equal(s.T(), 1, rawEvent.Version)

	var payload events.UserCreatedEvent
	err = events.ParsePayload(rawEvent, &payload)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "kafka_user4", payload.Username)
	assert.Equal(s.T(), "kafka4@example.com", payload.Email)
}

func (s *KafkaOutboxTestSuite) TestRegisterEachUserHasUniquePartitionKey() {
	email1 := "partition1@example.com"
	email2 := "partition2@example.com"
	email3 := "partition3@example.com"

	for idx, email := range []string{email1, email2, email3} {
		err := s.service.Register(s.ctx, model.RegisterInput{
			Username: fmt.Sprintf("user_partitioned_%d", idx),
			Email:    email,
			Password: "Secure123!",
		})
		assert.NoError(s.T(), err)
	}

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Len(evts, 3)

	keySet := make(map[string]bool)
	for _, e := range evts {
		if !keySet[e.Key] {
			keySet[e.Key] = true
		} else {
			s.FailNow("partition keys should be unique per user")
		}
	}
}

func (s *KafkaOutboxTestSuite) TestRegisterEachEventHasUniqueID() {
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("uniqueid_%d@example.com", i)
		err := s.service.Register(s.ctx, model.RegisterInput{
			Username: fmt.Sprintf("user_unique_%d", i),
			Email:    email,
			Password: "Secure123!",
		})
		assert.NoError(s.T(), err)
	}

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Len(evts, 5)

	idSet := make(map[string]bool)
	for _, e := range evts {
		assert.False(s.T(), idSet[e.ID.String()], "duplicated outbox event ID detected")
		idSet[e.ID.String()] = true
	}
}

func (s *KafkaOutboxTestSuite) TestRegisterWithInvalidEmail_NotWrittenToOutbox() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "invalid_kafka",
		Email:    "not-an-email",
		Password: "Secure123!",
	})
	assert.Error(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Empty(evts)
}

func (s *KafkaOutboxTestSuite) TestRegisterWithInvalidPassword_NotWrittenToOutbox() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "invalid_kafka2",
		Email:    "pass@example.com",
		Password: "short1!",
	})
	assert.Error(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Empty(evts)
}

func (s *KafkaOutboxTestSuite) TestRegisterWithDuplicateEmail_NotWrittenToOutbox() {
	email := "dup@example.com"

	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "user_dup1",
		Email:    email,
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	err = s.service.Register(s.ctx, model.RegisterInput{
		Username: "user_dup2",
		Email:    email,
		Password: "Secure123!",
	})
	assert.ErrorIs(s.T(), err, userDomain.ErrUserAlreadyExists)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Len(evts, 1)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventCreatedAtIsSet() {
	before := time.Now()
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_createdat",
		Email:    "createdat@example.com",
		Password: "Secure123!",
	})
	after := time.Now()
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Require().Len(evts, 1)

	assert.True(s.T(), !evts[0].CreatedAt.Before(before))
	assert.True(s.T(), evts[0].CreatedAt.Before(after.Add(time.Second)))
}

func (s *KafkaOutboxTestSuite) TestRegister_EventPayloadContainsUserID() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_payload_user",
		Email:    "payload@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Require().Len(evts, 1)

	user, err := s.userRepo.GetByEmail(s.ctx, "payload@example.com")
	s.Require().NoError(err)

	// Outbox key is the user ID used for Kafka partitioning
	s.Equal(user.ID().String(), evts[0].Key)

	var envelope events.Event
	err = json.Unmarshal(evts[0].Payload, &envelope)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), uuid.Nil, envelope.EventID)
	assert.Equal(s.T(), "user.created", envelope.EventType)
	assert.Greater(s.T(), envelope.Version, 0)

	var payload events.UserCreatedEvent
	err = json.Unmarshal(envelope.Payload, &payload)
	assert.NoError(s.T(), err)

	// Payload must contain the created user's ID
	s.Equal(user.ID(), payload.UserID)

	// Event ID must match the outbox row ID
	s.Equal(evts[0].ID, envelope.EventID)

	s.Equal("kafka_payload_user", payload.Username)
	s.Equal("payload@example.com", payload.Email)
}

func (s *KafkaOutboxTestSuite) TestRegister_EventPayloadContainsCreatedAt() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_payload_created",
		Email:    "created_evt@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Require().Len(evts, 1)

	var envelope events.Event
	err = json.Unmarshal(evts[0].Payload, &envelope)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), envelope.OccurredAt)

	var payload events.UserCreatedEvent
	err = json.Unmarshal(envelope.Payload, &payload)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), payload.CreatedAt)
}

func (s *KafkaOutboxTestSuite) TestRegister_ProcessedFieldIsNil() {
	err := s.service.Register(s.ctx, model.RegisterInput{
		Username: "kafka_processed",
		Email:    "processed@example.com",
		Password: "Secure123!",
	})
	assert.NoError(s.T(), err)

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 10)
	s.Require().Len(evts, 1)
	s.Nil(evts[0].ProcessedAt)
}

func (s *KafkaOutboxTestSuite) TestBatchRegistrations_AllStoredCorrectly() {
	registerCount := 20
	for i := range registerCount {
		email := fmt.Sprintf("batch_%d@example.com", i)
		err := s.service.Register(s.ctx, model.RegisterInput{
			Username: fmt.Sprintf("batchuser%d", i),
			Email:    email,
			Password: "Secure123!",
		})
		assert.NoError(s.T(), err)
	}

	evts, _ := s.outbox.GetUnprocessed(s.ctx, 50)
	s.Len(evts, registerCount)

	for _, evt := range evts {
		s.Equal("user.created", evt.Type)
		s.NotEmpty(evt.Key)
		assert.NotNil(s.T(), evt.Payload)
	}
}

func TestKafkaOutboxTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaOutboxTestSuite))
}
