package request_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	metricsmem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/metrics"
	transformationmem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/transformation"
	txmanagermem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/txmanager"
	"github.com/robertd2000/go-image-processing-app/processor/internal/usecase/request"
	"github.com/robertd2000/go-image-processing-app/processor/internal/usecase/request/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type RequestTransformationService interface {
	Execute(ctx context.Context, cmd model.Command) (*model.Result, error)
}
type RequestTransformationSuite struct {
	suite.Suite

	ctx context.Context

	repo      *transformationmem.InMemoryRepository
	txManager *txmanagermem.FakeTxManager

	service RequestTransformationService
}

func (s *RequestTransformationSuite) SetupTest() {
	s.ctx = context.Background()

	s.repo = transformationmem.NewInMemoryRepository()
	s.txManager = txmanagermem.NewFakeTxManager()

	logger := zap.NewNop().Sugar()
	metrics := metricsmem.NewFakeMetrics()

	s.service = request.NewRequestService(
		s.repo,
		s.txManager,
		logger,
		metrics,
	)
}

func (s *RequestTransformationSuite) newCommand() model.Command {
	source, err := transformDomain.NewSourceImage(
		"images/original/test.jpg",
		"image/jpeg",
		1920,
		1080,
	)
	require.NoError(s.T(), err)

	spec := transformDomain.TransformSpec{
		Operations: []transformDomain.Operation{
			{
				Resize: &transformDomain.ResizeParameters{
					Width:  400,
					Height: 300,
				},
			},
		},
	}

	return model.Command{
		ImageID: uuid.New(),
		Source:  source,
		Spec:    spec,
	}
}

func (s *RequestTransformationSuite) TestExecute_Create() {
	cmd := s.newCommand()

	result, err := s.service.Execute(s.ctx, cmd)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.NotEqual(uuid.Nil, result.ID)
	s.Equal(cmd.ImageID, result.ImageID)
	s.Equal(transformDomain.StatusPending, result.Status)

	tx := s.txManager.LastTx()
	s.Require().NotNil(tx)
	s.True(tx.Committed())
	s.False(tx.RolledBack())

	tr, err := s.repo.GetByID(s.ctx, result.ID)

	s.Require().NoError(err)
	s.Equal(result.ID, tr.ID())
	s.Equal(cmd.ImageID, tr.ImageID())
	s.Equal(transformDomain.StatusPending, tr.Status())
}

func (s *RequestTransformationSuite) TestExecute_ReturnExistingTransformation() {
	cmd := s.newCommand()

	first, err := s.service.Execute(s.ctx, cmd)
	s.Require().NoError(err)

	second, err := s.service.Execute(s.ctx, cmd)
	s.Require().NoError(err)

	s.Equal(first.ID, second.ID)
	s.Equal(first.Status, second.Status)
}

func (s *RequestTransformationSuite) TestExecute_InvalidCommand() {
	cmd := s.newCommand()
	cmd.ImageID = uuid.Nil

	result, err := s.service.Execute(s.ctx, cmd)

	s.Nil(result)
	s.Error(err)
}

func (s *RequestTransformationSuite) TestExecute_CommitTransaction() {
	cmd := s.newCommand()

	_, err := s.service.Execute(s.ctx, cmd)

	s.Require().NoError(err)

	tx := s.txManager.LastTx()

	s.Require().NotNil(tx)
	s.True(tx.Committed())
	s.False(tx.RolledBack())
}

func TestRequestTransformationSuite(t *testing.T) {
	suite.Run(t, new(RequestTransformationSuite))
}
