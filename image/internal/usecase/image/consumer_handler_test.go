package image

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"testing"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	imageInfra "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	jobmem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/job"
	storagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/storage"
	txmanagermem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/txmanager"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerHandler_Completed(t *testing.T) {
	imageRepo := imagemem.NewInMemoryImageRepo()
	storage := storagemem.NewInMemoryStorage()
	jobRepo := jobmem.NewInMemoryJobRepo()
	txManager := txmanagermem.NewFakeTxManager()
	extractor := imageInfra.NewMetadataExtractor()

	svc := NewImageService(imageRepo, storage, extractor, txManager, WithJobRepo(jobRepo))

	userID := uuid.New()
	img := createTestImage(t, svc, userID)

	handler := NewProcessingResultHandler(svc)

	eventID := uuid.New()
	ev := events.ImageProcessingCompleted{EventID: eventID, ImageID: img}
	payload, _ := json.Marshal(ev)
	msg := port.Message{
		Key:   img.String(),
		Value: payload,
		Headers: map[string]string{
			"event_type": events.EventTypeImageProcessingCompleted,
		},
	}

	err := handler(context.Background(), msg)
	require.NoError(t, err)

	entity, err := imageRepo.GetByID(context.Background(), img)
	require.NoError(t, err)
	assert.Equal(t, imageDomain.StatusCompleted, entity.Status())
}

func TestConsumerHandler_Failed(t *testing.T) {
	imageRepo := imagemem.NewInMemoryImageRepo()
	storage := storagemem.NewInMemoryStorage()
	jobRepo := jobmem.NewInMemoryJobRepo()
	txManager := txmanagermem.NewFakeTxManager()
	extractor := imageInfra.NewMetadataExtractor()

	svc := NewImageService(imageRepo, storage, extractor, txManager, WithJobRepo(jobRepo))

	userID := uuid.New()
	img := createTestImage(t, svc, userID)

	handler := NewProcessingResultHandler(svc)

	eventID := uuid.New()
	ev := events.ImageProcessingFailed{EventID: eventID, ImageID: img, Reason: "timeout"}
	payload, _ := json.Marshal(ev)
	msg := port.Message{
		Key:   img.String(),
		Value: payload,
		Headers: map[string]string{
			"event_type": events.EventTypeImageProcessingFailed,
		},
	}

	err := handler(context.Background(), msg)
	require.NoError(t, err)

	entity, err := imageRepo.GetByID(context.Background(), img)
	require.NoError(t, err)
	assert.Equal(t, imageDomain.StatusFailed, entity.Status())
}

func TestConsumerHandler_UnknownEventType(t *testing.T) {
	imageRepo := imagemem.NewInMemoryImageRepo()
	storage := storagemem.NewInMemoryStorage()
	jobRepo := jobmem.NewInMemoryJobRepo()
	txManager := txmanagermem.NewFakeTxManager()
	extractor := imageInfra.NewMetadataExtractor()

	svc := NewImageService(imageRepo, storage, extractor, txManager, WithJobRepo(jobRepo))

	handler := NewProcessingResultHandler(svc)

	msg := port.Message{
		Key:   uuid.New().String(),
		Value: []byte(`{}`),
		Headers: map[string]string{
			"event_type": "UnknownEvent",
		},
	}

	// should not error
	err := handler(context.Background(), msg)
	require.NoError(t, err)
}

func TestConsumerHandler_InvalidPayload(t *testing.T) {
	svc := NewImageService(nil, nil, nil, nil)

	handler := NewProcessingResultHandler(svc)

	msg := port.Message{
		Key:   uuid.New().String(),
		Value: []byte(`not-json`),
		Headers: map[string]string{
			"event_type": events.EventTypeImageProcessingCompleted,
		},
	}

	err := handler(context.Background(), msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestConsumerHandler_IdempotentCompleted(t *testing.T) {
	imageRepo := imagemem.NewInMemoryImageRepo()
	storage := storagemem.NewInMemoryStorage()
	jobRepo := jobmem.NewInMemoryJobRepo()
	txManager := txmanagermem.NewFakeTxManager()
	extractor := imageInfra.NewMetadataExtractor()

	svc := NewImageService(imageRepo, storage, extractor, txManager, WithJobRepo(jobRepo))

	userID := uuid.New()
	img := createTestImage(t, svc, userID)

	handler := NewProcessingResultHandler(svc)

	eventID := uuid.New()
	ev := events.ImageProcessingCompleted{EventID: eventID, ImageID: img}
	payload, _ := json.Marshal(ev)
	msg := port.Message{
		Key:   img.String(),
		Value: payload,
		Headers: map[string]string{
			"event_type": events.EventTypeImageProcessingCompleted,
		},
	}

	// first call
	err := handler(context.Background(), msg)
	require.NoError(t, err)

	// second call with same eventID — no-op
	err = handler(context.Background(), msg)
	require.NoError(t, err)

	entity, err := imageRepo.GetByID(context.Background(), img)
	require.NoError(t, err)
	assert.Equal(t, imageDomain.StatusCompleted, entity.Status())
}

func createTestImage(t *testing.T, svc *imageService, userID uuid.UUID) uuid.UUID {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	require.NoError(t, err)

	output, err := svc.UploadImage(context.Background(), model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     int64(buf.Len()),
		Reader:   bytes.NewReader(buf.Bytes()),
	})
	require.NoError(t, err)
	return output.ImageID
}
