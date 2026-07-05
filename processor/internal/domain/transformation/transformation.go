package transformation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Transformation struct {
	id uuid.UUID

	imageID uuid.UUID

	source SourceImage

	spec TransformSpec
	hash string

	status Status

	result *ResultImage

	errorMessage string

	startedAt   *time.Time
	completedAt *time.Time

	createdAt time.Time
	updatedAt time.Time
}

func NewTransformation(
	imageID uuid.UUID,
	source SourceImage,
	spec TransformSpec,
) (*Transformation, error) {

	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}

	now := time.Now().UTC()

	return &Transformation{
		id:        uuid.New(),
		imageID:   imageID,
		source:    source,
		spec:      spec,
		hash:      computeHash(imageID, spec),
		status:    StatusPending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreTransformation(
	id uuid.UUID,
	imageID uuid.UUID,
	source SourceImage,
	spec TransformSpec,
	hash string,
	status Status,
	result *ResultImage,
	errorMessage string,
	startedAt *time.Time,
	completedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*Transformation, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidTransformationID
	}

	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}

	if spec.Validate() != nil {
		return nil, ErrInvalidSpec
	}

	return &Transformation{
		id:           id,
		imageID:      imageID,
		spec:         spec,
		hash:         hash,
		status:       status,
		result:       result,
		errorMessage: errorMessage,
		startedAt:    startedAt,
		completedAt:  completedAt,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func computeHash(imageID uuid.UUID, spec TransformSpec) string {
	h := sha256.New()

	h.Write(imageID[:])

	data, err := json.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("marshal transform spec: %w", err))
	}

	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}
func (t *Transformation) Start() error {
	if t.status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()

	t.status = StatusProcessing
	t.startedAt = &now
	t.updatedAt = now

	return nil
}

func (t *Transformation) Complete(result ResultImage) error {
	if t.status != StatusProcessing {
		return ErrInvalidStatusTransition
	}

	if err := result.Validate(); err != nil {
		return err
	}

	now := time.Now().UTC()

	t.status = StatusCompleted
	t.result = &result
	t.errorMessage = ""
	t.completedAt = &now
	t.updatedAt = now

	return nil
}

func (t *Transformation) Fail(message string) error {
	if t.status != StatusProcessing {
		return ErrInvalidStatusTransition
	}

	if message == "" {
		return ErrInvalidErrorMessage
	}

	now := time.Now().UTC()

	t.status = StatusFailed
	t.errorMessage = message
	t.completedAt = &now
	t.updatedAt = now

	return nil
}

func (t *Transformation) Duration() time.Duration {
	if t.startedAt == nil || t.completedAt == nil {
		return 0
	}

	return t.completedAt.Sub(*t.startedAt)
}

func (t *Transformation) ID() uuid.UUID {
	return t.id
}

func (t *Transformation) ImageID() uuid.UUID {
	return t.imageID
}

func (t *Transformation) Spec() TransformSpec {
	return t.spec
}

func (t *Transformation) Hash() string {
	return t.hash
}

func (t *Transformation) Status() Status {
	return t.status
}

func (t *Transformation) Result() *ResultImage {
	return t.result
}

func (t *Transformation) ErrorMessage() string {
	return t.errorMessage
}

func (t *Transformation) StartedAt() *time.Time {
	return t.startedAt
}

func (t *Transformation) CompletedAt() *time.Time {
	return t.completedAt
}

func (t *Transformation) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Transformation) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Transformation) String() string {
	spec, _ := json.Marshal(t.spec)

	specStr := string(spec)
	if len(specStr) > 120 {
		specStr = specStr[:120] + "..."
	}

	return fmt.Sprintf(
		"Transformation{id=%s,imageID=%s,status=%s,hash=%s,spec=%s}",
		t.id,
		t.imageID,
		t.status,
		t.hash,
		specStr,
	)
}

func (t *Transformation) Source() SourceImage {
	return t.source
}
