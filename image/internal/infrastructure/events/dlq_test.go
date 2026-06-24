package events

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dlqPublisher struct {
	mu       sync.Mutex
	messages []port.Message
	topics   []string
}

func (p *dlqPublisher) Publish(_ context.Context, topic string, msg port.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, msg)
	p.topics = append(p.topics, topic)
	return nil
}

func (p *dlqPublisher) Close() error { return nil }

func TestWithDLQ_Success_DoesNotPublish(t *testing.T) {
	pub := &dlqPublisher{}

	handler := WithDLQ(
		func(_ context.Context, _ port.Message) error { return nil },
		pub, "dlq-topic",
	)

	err := handler(context.Background(), port.Message{Key: "1", Value: []byte("ok")})
	require.NoError(t, err)

	pub.mu.Lock()
	assert.Len(t, pub.messages, 0)
	pub.mu.Unlock()
}

func TestWithDLQ_Error_PublishesToDLQ(t *testing.T) {
	pub := &dlqPublisher{}

	handler := WithDLQ(
		func(_ context.Context, _ port.Message) error { return errors.New("processing failed") },
		pub, "dlq-topic",
	)

	msg := port.Message{
		Key:   "img-1",
		Value: []byte(`{"image_id":"abc"}`),
		Headers: map[string]string{
			"event_type": "ImageProcessingCompleted",
		},
	}

	err := handler(context.Background(), msg)
	require.NoError(t, err)

	pub.mu.Lock()
	require.Len(t, pub.messages, 1)
	assert.Equal(t, "dlq-topic", pub.topics[0])
	assert.Equal(t, msg.Key, pub.messages[0].Key)
	assert.Equal(t, msg.Value, pub.messages[0].Value)
	assert.Equal(t, "processing failed", pub.messages[0].Headers["dlq_error"])
	assert.NotEmpty(t, pub.messages[0].Headers["dlq_timestamp"])
	assert.Equal(t, "ImageProcessingCompleted", pub.messages[0].Headers["event_type"])
	pub.mu.Unlock()
}

func TestWithDLQ_PublishError_DoesNotPanic(t *testing.T) {
	failingPub := &fakePublisher{fail: true}

	handler := WithDLQ(
		func(_ context.Context, _ port.Message) error { return errors.New("err") },
		failingPub, "dlq-topic",
	)

	err := handler(context.Background(), port.Message{Key: "1"})
	require.NoError(t, err)
}
