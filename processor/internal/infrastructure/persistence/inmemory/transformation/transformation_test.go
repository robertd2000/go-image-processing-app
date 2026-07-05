package transformationmem_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	transformDomain "github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
	transformationmem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/transformation"
	txmanagermem "github.com/robertd2000/go-image-processing-app/processor/internal/infrastructure/persistence/inmemory/txmanager"
)

type RepositorySuite struct {
	suite.Suite

	ctx  context.Context
	repo *transformationmem.InMemoryRepository
	tx   *txmanagermem.FakeTx
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func (s *RepositorySuite) SetupTest() {
	s.ctx = context.Background()
	s.repo = transformationmem.NewInMemoryRepository()
	s.tx = &txmanagermem.FakeTx{}
}

func (s *RepositorySuite) newTransformation() *transformDomain.Transformation {
	source, err := transformDomain.NewSourceImage(
		"images/test.jpg",
		"image/jpeg",
		800,
		600,
	)
	require.NoError(s.T(), err)

	spec := transformDomain.TransformSpec{
		Operations: []transformDomain.Operation{
			{
				Resize: &transformDomain.ResizeParameters{
					Width:         400,
					Height:        300,
					MaintainRatio: true,
					Quality:       90,
				},
			},
		},
	}

	tr, err := transformDomain.NewTransformation(
		uuid.New(),
		source,
		spec,
	)

	require.NoError(s.T(), err)

	return tr
}

func (s *RepositorySuite) TestCreate() {
	tr := s.newTransformation()

	err := s.repo.Create(s.ctx, s.tx, tr)

	require.NoError(s.T(), err)

	got, err := s.repo.GetByID(s.ctx, tr.ID())

	require.NoError(s.T(), err)

	assert.Equal(s.T(), tr.ID(), got.ID())
	assert.Equal(s.T(), tr.Hash(), got.Hash())
}

func (s *RepositorySuite) TestGetByID() {
	tr := s.newTransformation()

	require.NoError(s.T(),
		s.repo.Create(s.ctx, s.tx, tr),
	)

	got, err := s.repo.GetByID(s.ctx, tr.ID())

	require.NoError(s.T(), err)

	assert.Equal(s.T(), tr.ID(), got.ID())
}

func (s *RepositorySuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(s.ctx, uuid.New())

	require.ErrorIs(s.T(), err, transformDomain.ErrNotFound)
}

func (s *RepositorySuite) TestGetByImageAndHash() {
	tr := s.newTransformation()

	require.NoError(s.T(),
		s.repo.Create(s.ctx, s.tx, tr),
	)

	got, err := s.repo.GetByImageAndHash(
		s.ctx,
		tr.ImageID(),
		tr.Hash(),
	)

	require.NoError(s.T(), err)

	assert.Equal(s.T(), tr.ID(), got.ID())
}

func (s *RepositorySuite) TestUpdate() {
	tr := s.newTransformation()

	require.NoError(s.T(),
		s.repo.Create(s.ctx, s.tx, tr),
	)

	require.NoError(s.T(), tr.Start())

	require.NoError(s.T(),
		s.repo.Update(s.ctx, s.tx, tr),
	)

	got, err := s.repo.GetByID(s.ctx, tr.ID())

	require.NoError(s.T(), err)

	assert.Equal(
		s.T(),
		transformDomain.StatusProcessing,
		got.Status(),
	)
}

func (s *RepositorySuite) TestAcquireNextPending() {
	tr := s.newTransformation()

	require.NoError(s.T(),
		s.repo.Create(s.ctx, s.tx, tr),
	)

	got, err := s.repo.AcquireNextPending(
		s.ctx,
		s.tx,
	)

	require.NoError(s.T(), err)

	assert.Equal(s.T(), tr.ID(), got.ID())
}

func (s *RepositorySuite) TestAcquireNextPending_NoPending() {
	_, err := s.repo.AcquireNextPending(
		s.ctx,
		s.tx,
	)

	require.ErrorIs(s.T(), err, transformDomain.ErrNotFound)
}

func (s *RepositorySuite) TestAcquireNextPending_SkipLocked() {
	tr := s.newTransformation()

	require.NoError(s.T(),
		s.repo.Create(s.ctx, s.tx, tr),
	)

	_, err := s.repo.AcquireNextPending(
		s.ctx,
		s.tx,
	)

	require.NoError(s.T(), err)

	tx2 := &txmanagermem.FakeTx{}

	_, err = s.repo.AcquireNextPending(
		s.ctx,
		tx2,
	)

	require.ErrorIs(s.T(), err, transformDomain.ErrNotFound)
}
