package image_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/domain/events"
	imageInfra "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/image"
	imagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/image"
	jobmem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/job"
	outboxmem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/outbox"
	storagemem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/storage"
	txmanagermem "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/inmemory/txmanager"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	imageUsecase "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	txtx "github.com/robertd2000/go-image-processing-app/image/internal/domain/tx"
	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/image/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ImageService interface {
	UploadImage(ctx context.Context, input model.UploadImageInput) (*model.UploadImageOutput, error)
	GetImage(ctx context.Context, imageID uuid.UUID) (*model.ImageOutput, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
	ListImages(ctx context.Context, input model.ListImagesInput) (*model.ListImagesOutput, error)
	HandleImageProcessed(ctx context.Context, eventID, imageID uuid.UUID) error
	HandleImageProcessingFailed(ctx context.Context, eventID, imageID uuid.UUID, reason string) error
}

type imageServiceTestSuite struct {
	suite.Suite

	ctx context.Context

	service   ImageService
	imageRepo imageDomain.Repository
	storage   port.Storage

	txManager port.TxManager

	metadataExtractor port.Extractor
}

func (s *imageServiceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.imageRepo = imagemem.NewInMemoryImageRepo()
	s.storage = storagemem.NewInMemoryStorage()
	s.metadataExtractor = imageInfra.NewMetadataExtractor()
	s.txManager = txmanagermem.NewFakeTxManager()

	s.service = imageUsecase.NewImageService(s.imageRepo, s.storage, s.metadataExtractor, s.txManager)
}

// SUCCESS

func (s *imageServiceTestSuite) TestUploadImage_Success() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := s.service.UploadImage(s.ctx, input)

	if !assert.NoError(s.T(), err) {
		return
	}

	assert.NotNil(s.T(), output)
	assert.NotEqual(s.T(), uuid.Nil, output.ImageID)
	assert.WithinDuration(s.T(), time.Now(), output.CreatedAt, time.Second)
}

// VALIDATION

func (s *imageServiceTestSuite) TestUploadImage_InvalidUserID() {
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   uuid.Nil,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidUserID)
}

func (s *imageServiceTestSuite) TestUploadImage_NoReader() {
	input := model.UploadImageInput{
		UserID: uuid.New(),
		Size:   100,
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidImageMissingReader)
}

func (s *imageServiceTestSuite) TestUploadImage_InvalidSize() {
	buf, _ := generateTestImage()

	input := model.UploadImageInput{
		UserID: uuid.New(),
		Reader: bytes.NewReader(buf.Bytes()),
		Size:   0,
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.ErrorIs(s.T(), err, imageDomain.ErrInvalidImageSize)
}

// INVALID IMAGE

func (s *imageServiceTestSuite) TestUploadImage_InvalidImageData() {
	input := model.UploadImageInput{
		UserID: uuid.New(),
		Reader: bytes.NewReader([]byte("not-an-image")),
		Size:   int64(len("not-an-image")),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
}

// STORAGE ERROR

type failingStorage struct {
	port.Storage
}

func (f *failingStorage) Put(ctx context.Context, key string, r io.Reader, size int64, mime string) error {
	return errors.New("storage error")
}

func (s *imageServiceTestSuite) TestUploadImage_StorageFails() {
	userID := uuid.New()
	buf, size := generateTestImage()

	s.service = imageUsecase.NewImageService(
		s.imageRepo,
		&failingStorage{},
		s.metadataExtractor,
		s.txManager,
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "storage")
}

// EXTENSION

func (s *imageServiceTestSuite) TestUploadImage_KnownMime_IgnoresFilenameExtension() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.unknownext",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.NoError(s.T(), err)
}

func (s *imageServiceTestSuite) TestUploadImage_GIF() {
	userID := uuid.New()
	buf, size := generateTestGIF()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.gif",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := s.service.UploadImage(s.ctx, input)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), output)

	res, err := s.service.GetImage(s.ctx, output.ImageID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "image/gif", res.MimeType)
}

// LARGE FILE LIMIT
func (s *imageServiceTestSuite) TestUploadImage_TooLarge() {
	userID := uuid.New()

	large := bytes.Repeat([]byte("a"), 20<<20) // 20MB

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "big.png",
		Size:     int64(len(large)),
		Reader:   bytes.NewReader(large),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)
}

// REPO ERROR + ROLLBACK

type failingRepo struct {
	imageDomain.Repository
}

func (f *failingRepo) Save(ctx context.Context, tx txtx.Tx, img *imageDomain.Image) error {
	return errors.New("repo error")
}

func (s *imageServiceTestSuite) TestUploadImage_RepoFails_ShouldRollbackStorage() {
	userID := uuid.New()
	buf, size := generateTestImage()

	spy := storagemem.NewSpyStorage()

	s.service = imageUsecase.NewImageService(
		&failingRepo{},
		spy,
		s.metadataExtractor,
		s.txManager,
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)

	assert.Error(s.T(), err)

	assert.True(s.T(), spy.PutCalled)
	assert.True(s.T(), spy.DeleteCalled)

	txMgr := s.txManager.(*txmanagermem.FakeTxManager)
	tx := txMgr.LastTx()
	s.Require().NotNil(tx)
	s.True(tx.RolledBack())
	s.False(tx.Committed())
}

// TRANSACTION LIFECYCLE

func (s *imageServiceTestSuite) TestUploadImage_TransactionCommitted() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	_, err := s.service.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	txMgr := s.txManager.(*txmanagermem.FakeTxManager)
	tx := txMgr.LastTx()
	s.Require().NotNil(tx)
	s.True(tx.Committed())
	s.False(tx.RolledBack())
}

// GetImage tests

func (s *imageServiceTestSuite) TestGetImage_Success() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)
	s.Require().NotNil(uploadRes)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Equal(uploadRes.ImageID, res.ImageID)
	s.NotEmpty(res.URL)
	s.Equal("test.png", res.FileName)
	s.Equal(int64(size), res.Size)
	s.Equal("image/png", res.MimeType)
}

func (s *imageServiceTestSuite) TestGetImage_NotFound() {
	res, err := s.service.GetImage(s.ctx, uuid.New())

	s.Require().Error(err)
	s.Nil(res)
}

func (s *imageServiceTestSuite) TestGetImage_StorageError() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	err = s.storage.Delete(s.ctx, string(img.StorageKey()))
	s.Require().NoError(err)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)

	s.Require().Error(err)
	s.Nil(res)
}

func (s *imageServiceTestSuite) TestGetImage_MetadataMapping() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	res, err := s.service.GetImage(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	s.Equal(10, res.Width)
	s.Equal(10, res.Height)
}

// DeleteImage

func (s *imageServiceTestSuite) TestDeleteImage_Success() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().NoError(err)

	_, err = s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	s.ErrorIs(err, imageDomain.ErrNotFound)
}

func (s *imageServiceTestSuite) TestDeleteImage_NotFound() {
	err := s.service.DeleteImage(s.ctx, uuid.New())

	s.Require().Error(err)
}

func (s *imageServiceTestSuite) TestDeleteImage_StorageAlreadyDeleted() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	_ = s.storage.Delete(s.ctx, string(img.StorageKey()))

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().NoError(err)
}

type failingDeleteStorage struct {
	port.Storage
}

func (f *failingDeleteStorage) Delete(ctx context.Context, key string) error {
	return errors.New("storage delete error")
}

func (s *imageServiceTestSuite) TestDeleteImage_StorageFails() {
	userID := uuid.New()
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	s.service = imageUsecase.NewImageService(
		s.imageRepo,
		&failingDeleteStorage{Storage: s.storage},
		s.metadataExtractor,
		s.txManager,
	)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().Error(err)
	s.Contains(err.Error(), "storage")
}

func (s *imageServiceTestSuite) TestDeleteImage_Idempotent() {
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   uuid.New(),
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)
	s.Require().NoError(err)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().Error(err)
}

type failingDeleteRepo struct {
	imageDomain.Repository
}

func (f *failingDeleteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return errors.New("repo delete error")
}

func (s *imageServiceTestSuite) TestDeleteImage_RepoFails() {
	userID := uuid.New()
	buf, size := generateTestImage()

	uploadRes, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	s.service = imageUsecase.NewImageService(
		&failingDeleteRepo{Repository: s.imageRepo},
		s.storage,
		s.metadataExtractor,
		s.txManager,
	)

	err = s.service.DeleteImage(s.ctx, uploadRes.ImageID)

	s.Require().Error(err)
	s.Contains(err.Error(), "delete image")
}

// ListImages

func (s *imageServiceTestSuite) TestListImages_Success() {
	userID := uuid.New()

	for i := range 3 {
		buf, size := generateTestImage()

		_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
			UserID:   userID,
			Filename: fmt.Sprintf("img_%d.png", i),
			Reader:   buf,
			Size:     size,
		})
		s.Require().NoError(err)
	}

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Len(res.Items, 3)
	s.Equal(3, res.Total)

	for _, item := range res.Items {
		s.NotEmpty(item.URL)
		s.Equal(userID, item.UserID)
	}
}

func (s *imageServiceTestSuite) TestListImages_Empty() {
	userID := uuid.New()

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Len(res.Items, 0)
	s.Equal(0, res.Total)
}

func (s *imageServiceTestSuite) TestListImages_Pagination() {
	userID := uuid.New()

	for i := range 5 {
		buf, size := generateTestImage()

		_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
			UserID:   userID,
			Filename: fmt.Sprintf("img_%d.png", i),
			Reader:   buf,
			Size:     size,
		})
		s.Require().NoError(err)
	}

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  2,
		Offset: 0,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Len(res.Items, 2)
	s.Equal(5, res.Total)
}

func (s *imageServiceTestSuite) TestListImages_Offset() {
	userID := uuid.New()

	for i := range 5 {
		buf, size := generateTestImage()

		_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
			UserID:   userID,
			Filename: fmt.Sprintf("img_%d.png", i),
			Reader:   buf,
			Size:     size,
		})
		s.Require().NoError(err)
	}

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  2,
		Offset: 2,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)

	s.Len(res.Items, 2)
}

func (s *imageServiceTestSuite) TestListImages_DefaultLimit() {
	userID := uuid.New()

	for i := range 3 {
		buf, size := generateTestImage()

		_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
			UserID:   userID,
			Filename: fmt.Sprintf("img_%d.png", i),
			Reader:   buf,
			Size:     size,
		})
		s.Require().NoError(err)
	}

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  0,
		Offset: 0,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Len(res.Items, 3)
	s.Equal(10, res.Limit)
}

func (s *imageServiceTestSuite) TestListImages_MaxLimit() {
	userID := uuid.New()

	for i := range 3 {
		buf, size := generateTestImage()

		_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
			UserID:   userID,
			Filename: fmt.Sprintf("img_%d.png", i),
			Reader:   buf,
			Size:     size,
		})
		s.Require().NoError(err)
	}

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  150,
		Offset: 0,
	})

	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Len(res.Items, 3)
	s.Equal(100, res.Limit)
}

type brokenStorage struct {
	port.Storage
}

func (b *brokenStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("storage error")
}

func (s *imageServiceTestSuite) TestListImages_InvalidUserID() {
	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: uuid.Nil,
		Limit:  10,
		Offset: 0,
	})

	s.ErrorIs(err, imageDomain.ErrInvalidUserID)
	s.Nil(res)
}

func (s *imageServiceTestSuite) TestListImages_InvalidPagination() {
	userID := uuid.New()

	_, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  -1,
		Offset: 0,
	})
	s.ErrorIs(err, imageDomain.ErrInvalidPagination)

	_, err = s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  10,
		Offset: -1,
	})
	s.ErrorIs(err, imageDomain.ErrInvalidPagination)
}

func (s *imageServiceTestSuite) TestListImages_StorageError() {
	userID := uuid.New()

	buf, size := generateTestImage()

	_, err := s.service.UploadImage(s.ctx, model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Reader:   buf,
		Size:     size,
	})
	s.Require().NoError(err)

	s.service = imageUsecase.NewImageService(
		s.imageRepo,
		&brokenStorage{Storage: s.storage},
		s.metadataExtractor,
		s.txManager,
	)

	res, err := s.service.ListImages(s.ctx, model.ListImagesInput{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})

	s.Require().Error(err)
	s.Nil(res)
}

// EVENT TESTS

func (s *imageServiceTestSuite) TestUploadImage_StatusPending() {
	userID := uuid.New()
	buf, size := generateTestImage()

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := s.service.UploadImage(s.ctx, input)
	s.Require().NoError(err)
	s.Equal("pending", output.Status)
}

func (s *imageServiceTestSuite) TestUploadImage_CreatesOutboxAndJob() {
	userID := uuid.New()
	buf, size := generateTestImage()

	outboxRepo := outboxmem.NewInMemoryOutboxRepo()
	jobRepo := jobmem.NewInMemoryJobRepo()

	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
		imageUsecase.WithOutbox(outboxRepo),
		imageUsecase.WithJobRepo(jobRepo),
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := svc.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	// verify outbox event
	pending, err := outboxRepo.FetchPending(s.ctx, 10)
	s.Require().NoError(err)
	s.Len(pending, 1)
	s.Equal(output.ImageID, pending[0].AggregateID)
	s.Equal(events.EventTypeImageUploaded, pending[0].EventType)

	var ev events.ImageUploaded
	err = json.Unmarshal(pending[0].Payload, &ev)
	s.Require().NoError(err)
	s.Equal(output.ImageID, ev.ImageID)
	s.Equal(userID, ev.UserID)

	// verify job: MarkCompleted returns true first time
	ok, err := jobRepo.MarkCompleted(s.ctx, output.ImageID, uuid.New())
	s.Require().NoError(err)
	s.True(ok)

	// verify idempotent: MarkCompleted returns false second time
	ok, err = jobRepo.MarkCompleted(s.ctx, output.ImageID, uuid.New())
	s.Require().NoError(err)
	s.False(ok)
}

func (s *imageServiceTestSuite) TestHandleImageProcessed_Success() {
	userID := uuid.New()
	buf, size := generateTestImage()

	jobRepo := jobmem.NewInMemoryJobRepo()

	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
		imageUsecase.WithJobRepo(jobRepo),
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := svc.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	eventID := uuid.New()
	err = svc.HandleImageProcessed(s.ctx, eventID, output.ImageID)
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal(imageDomain.StatusCompleted, img.Status())

	// verify via GetImage as well
	res, err := svc.GetImage(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal("completed", res.Status)
}

func (s *imageServiceTestSuite) TestHandleImageProcessingFailed_Success() {
	userID := uuid.New()
	buf, size := generateTestImage()

	jobRepo := jobmem.NewInMemoryJobRepo()

	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
		imageUsecase.WithJobRepo(jobRepo),
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := svc.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	eventID := uuid.New()
	err = svc.HandleImageProcessingFailed(s.ctx, eventID, output.ImageID, "processing error")
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal(imageDomain.StatusFailed, img.Status())

	res, err := svc.GetImage(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal("failed", res.Status)
}

func (s *imageServiceTestSuite) TestHandleImageProcessed_Idempotent() {
	userID := uuid.New()
	buf, size := generateTestImage()

	jobRepo := jobmem.NewInMemoryJobRepo()

	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
		imageUsecase.WithJobRepo(jobRepo),
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := svc.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	// first call succeeds
	eventID := uuid.New()
	err = svc.HandleImageProcessed(s.ctx, eventID, output.ImageID)
	s.Require().NoError(err)

	// second call (same eventID) should be no-op
	err = svc.HandleImageProcessed(s.ctx, eventID, output.ImageID)
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal(imageDomain.StatusCompleted, img.Status())
}

func (s *imageServiceTestSuite) TestHandleImageProcessingFailed_Idempotent() {
	userID := uuid.New()
	buf, size := generateTestImage()

	jobRepo := jobmem.NewInMemoryJobRepo()

	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
		imageUsecase.WithJobRepo(jobRepo),
	)

	input := model.UploadImageInput{
		UserID:   userID,
		Filename: "test.png",
		Size:     size,
		Reader:   bytes.NewReader(buf.Bytes()),
	}

	output, err := svc.UploadImage(s.ctx, input)
	s.Require().NoError(err)

	// first call succeeds
	eventID := uuid.New()
	err = svc.HandleImageProcessingFailed(s.ctx, eventID, output.ImageID, "err")
	s.Require().NoError(err)

	// second call (same image, diff eventID) should be no-op since already failed
	err = svc.HandleImageProcessingFailed(s.ctx, uuid.New(), output.ImageID, "err again")
	s.Require().NoError(err)

	img, err := s.imageRepo.GetByID(s.ctx, output.ImageID)
	s.Require().NoError(err)
	s.Equal(imageDomain.StatusFailed, img.Status())
}

func (s *imageServiceTestSuite) TestHandleImageProcessed_WithoutJobRepo() {
	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
	)

	err := svc.HandleImageProcessed(s.ctx, uuid.New(), uuid.New())
	s.Require().NoError(err)
}

func (s *imageServiceTestSuite) TestHandleImageProcessingFailed_WithoutJobRepo() {
	svc := imageUsecase.NewImageService(
		s.imageRepo, s.storage, s.metadataExtractor, s.txManager,
	)

	err := svc.HandleImageProcessingFailed(s.ctx, uuid.New(), uuid.New(), "reason")
	s.Require().NoError(err)
}

// HELPERS

func generateTestImage() (*bytes.Buffer, int64) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)

	return buf, int64(buf.Len())
}

func generateTestGIF() (*bytes.Buffer, int64) {
	palette := color.Palette{color.Black, color.White}
	img := image.NewPaletted(image.Rect(0, 0, 5, 5), palette)

	buf := new(bytes.Buffer)
	_ = gif.Encode(buf, img, nil)

	return buf, int64(buf.Len())
}

func TestImageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(imageServiceTestSuite))
}
