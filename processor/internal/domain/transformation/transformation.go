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
	id      uuid.UUID
	imageID uuid.UUID

	storageKey string
	mimeType   string
	width      int
	height     int

	spec json.RawMessage
	hash string

	status       Status
	resultKey    string
	errorMessage string

	startedAt   *time.Time
	completedAt *time.Time
	createdAt   time.Time
}

func NewTransformation(
	imageID uuid.UUID,
	storageKey string,
	mimeType string,
	width, height int,
	spec json.RawMessage,
) (*Transformation, error) {

	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}

	if storageKey == "" {
		return nil, ErrInvalidStorageKey
	}

	if mimeType == "" {
		return nil, ErrInvalidMimeType
	}

	if width <= 0 || height <= 0 {
		return nil, ErrInvalidImageSize
	}

	if len(spec) == 0 || !json.Valid(spec) {
		return nil, ErrInvalidSpec
	}

	return &Transformation{
		id:         uuid.New(),
		imageID:    imageID,
		storageKey: storageKey,
		mimeType:   mimeType,
		width:      width,
		height:     height,
		spec:       spec,
		hash:       computeHash(imageID, spec),
		status:     StatusPending,
		createdAt:  time.Now().UTC(),
	}, nil
}

func RestoreTransformation(
	id uuid.UUID,
	imageID uuid.UUID,
	storageKey string,
	mimeType string,
	width, height int,
	spec json.RawMessage,
	hash string,
	status Status,
	resultKey string,
	errorMessage string,
	startedAt *time.Time,
	completedAt *time.Time,
	createdAt time.Time,
) (*Transformation, error) {

	if id == uuid.Nil {
		return nil, ErrInvalidTransformationID
	}

	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}

	return &Transformation{
		id:           id,
		imageID:      imageID,
		storageKey:   storageKey,
		mimeType:     mimeType,
		width:        width,
		height:       height,
		spec:         spec,
		hash:         hash,
		status:       status,
		resultKey:    resultKey,
		errorMessage: errorMessage,
		startedAt:    startedAt,
		completedAt:  completedAt,
		createdAt:    createdAt,
	}, nil
}

func computeHash(imageID uuid.UUID, spec json.RawMessage) string {
	h := sha256.New()

	h.Write(imageID[:])
	h.Write(spec)

	return hex.EncodeToString(h.Sum(nil))
}

func (t *Transformation) Start() error {
	if t.status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()

	t.status = StatusProcessing
	t.startedAt = &now

	return nil
}

func (t *Transformation) Complete(resultKey string) error {
	if t.status != StatusProcessing {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()

	t.status = StatusDone
	t.resultKey = resultKey
	t.completedAt = &now
	t.errorMessage = ""

	return nil
}

func (t *Transformation) Fail(message string) error {
	if t.status != StatusProcessing {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()

	t.status = StatusFailed
	t.errorMessage = message
	t.completedAt = &now

	return nil
}

func (t *Transformation) ID() uuid.UUID {
	return t.id
}

func (t *Transformation) ImageID() uuid.UUID {
	return t.imageID
}

func (t *Transformation) StorageKey() string {
	return t.storageKey
}

func (t *Transformation) MimeType() string {
	return t.mimeType
}

func (t *Transformation) Width() int {
	return t.width
}

func (t *Transformation) Height() int {
	return t.height
}

func (t *Transformation) Spec() json.RawMessage {
	return t.spec
}

func (t *Transformation) Hash() string {
	return t.hash
}

func (t *Transformation) Status() Status {
	return t.status
}

func (t *Transformation) ResultKey() string {
	return t.resultKey
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

func (t *Transformation) Duration() time.Duration {
	if t.startedAt == nil || t.completedAt == nil {
		return 0
	}

	return t.completedAt.Sub(*t.startedAt)
}

func (t *Transformation) String() string {
	spec := string(t.spec)
	if len(spec) > 120 {
		spec = spec[:120] + "..."
	}

	return fmt.Sprintf(
		"Transformation{id=%s,image=%s,status=%s,hash=%s}",
		t.id,
		t.imageID,
		t.status,
		t.hash,
	)
}
