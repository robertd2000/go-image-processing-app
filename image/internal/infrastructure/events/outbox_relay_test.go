package events

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	domainEvents "github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	outboxmem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/outbox"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	mu       sync.Mutex
	messages []port.Message
	fail     bool
}

func (p *fakePublisher) Publish(_ context.Context, _ string, msg port.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return errors.New("publish error")
	}
	p.messages = append(p.messages, msg)
	return nil
}

func (p *fakePublisher) Close() error { return nil }

func TestOutboxRelay_PublishesPendingEvents(t *testing.T) {
	outboxRepo := outboxmem.NewInMemoryOutboxRepo()
	pub := &fakePublisher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	imageID := uuid.New()
	eventID := uuid.New()
	payload, _ := json.Marshal(domainEvents.ImageUploaded{EventID: eventID, ImageID: imageID})

	err := outboxRepo.Save(ctx, nil, &port.OutboxEvent{
		ID:          eventID,
		AggregateID: imageID,
		EventType:   domainEvents.EventTypeImageUploaded,
		Payload:     payload,
		Status:      port.OutboxStatusPending,
		CreatedAt:   time.Now(),
	})
	require.NoError(t, err)

	go RunOutboxRelay(ctx, outboxRepo, pub, "image-processing-requested", 50*time.Millisecond, 10)

	time.Sleep(200 * time.Millisecond)

	pub.mu.Lock()
	assert.Len(t, pub.messages, 1)
	msg := pub.messages[0]
	pub.mu.Unlock()

	assert.Equal(t, imageID.String(), msg.Key)
	assert.Equal(t, domainEvents.EventTypeImageUploaded, msg.Headers["event_type"])

	// verify event is no longer pending
	pending, err := outboxRepo.FetchPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 0)
}

func TestOutboxRelay_MarksFailedOnPublishError(t *testing.T) {
	outboxRepo := outboxmem.NewInMemoryOutboxRepo()
	pub := &fakePublisher{fail: true}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	imageID := uuid.New()
	eventID := uuid.New()
	payload, _ := json.Marshal(domainEvents.ImageUploaded{EventID: eventID, ImageID: imageID})

	err := outboxRepo.Save(ctx, nil, &port.OutboxEvent{
		ID:          eventID,
		AggregateID: imageID,
		EventType:   domainEvents.EventTypeImageUploaded,
		Payload:     payload,
		Status:      port.OutboxStatusPending,
		CreatedAt:   time.Now(),
	})
	require.NoError(t, err)

	go RunOutboxRelay(ctx, outboxRepo, pub, "image-processing-requested", 50*time.Millisecond, 10)

	time.Sleep(200 * time.Millisecond)

	// verify event is marked as failed
	pending, err := outboxRepo.FetchPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 0)
}

func TestOutboxRelay_RespectsContextCancellation(t *testing.T) {
	outboxRepo := outboxmem.NewInMemoryOutboxRepo()
	pub := &fakePublisher{}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		RunOutboxRelay(ctx, outboxRepo, pub, "test", 10*time.Hour, 10)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// relay exited cleanly
	case <-time.After(time.Second):
		t.Fatal("relay did not stop after context cancellation")
	}
}
