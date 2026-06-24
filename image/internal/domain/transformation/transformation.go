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
	id           uuid.UUID
	imageID      uuid.UUID
	spec         json.RawMessage
	hash         string
	status       Status
	resultKey    string
	errorMessage string
	startedAt    *time.Time
	completedAt  *time.Time
	duration     int64
	createdAt    time.Time
}

func NewTransformation(imageID uuid.UUID, spec json.RawMessage) (*Transformation, error) {
	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}
	if len(spec) == 0 || !json.Valid(spec) {
		return nil, ErrInvalidSpec
	}

	hash := computeHash(imageID, spec)

	return &Transformation{
		id:        uuid.New(),
		imageID:   imageID,
		spec:      spec,
		hash:      hash,
		status:    StatusPending,
		createdAt: time.Now(),
	}, nil
}

func RestoreTransformation(
	id, imageID uuid.UUID,
	spec json.RawMessage, hash string,
	status Status,
	resultKey, errorMessage string,
	startedAt, completedAt *time.Time,
	duration int64,
	createdAt time.Time,
) (*Transformation, error) {
	if imageID == uuid.Nil {
		return nil, ErrInvalidImageID
	}
	return &Transformation{
		id:           id,
		imageID:      imageID,
		spec:         spec,
		hash:         hash,
		status:       status,
		resultKey:    resultKey,
		errorMessage: errorMessage,
		startedAt:    startedAt,
		completedAt:  completedAt,
		duration:     duration,
		createdAt:    createdAt,
	}, nil
}

func computeHash(imageID uuid.UUID, spec json.RawMessage) string {
	h := sha256.Sum256(append(imageID[:], spec...))
	return hex.EncodeToString(h[:])
}

func (t *Transformation) ID() uuid.UUID           { return t.id }
func (t *Transformation) ImageID() uuid.UUID       { return t.imageID }
func (t *Transformation) Spec() json.RawMessage    { return t.spec }
func (t *Transformation) Hash() string             { return t.hash }
func (t *Transformation) Status() Status           { return t.status }
func (t *Transformation) ResultKey() string        { return t.resultKey }
func (t *Transformation) ErrorMessage() string     { return t.errorMessage }
func (t *Transformation) StartedAt() *time.Time    { return t.startedAt }
func (t *Transformation) CompletedAt() *time.Time  { return t.completedAt }
func (t *Transformation) Duration() int64          { return t.duration }
func (t *Transformation) CreatedAt() time.Time     { return t.createdAt }

func (t *Transformation) SetStatus(s Status)        { t.status = s }
func (t *Transformation) SetResultKey(k string)     { t.resultKey = k }
func (t *Transformation) SetErrorMessage(m string)  { t.errorMessage = m }

func (t *Transformation) String() string {
	specStr := string(t.spec)
	if len(specStr) > 100 {
		specStr = specStr[:100] + "..."
	}
	return fmt.Sprintf("Transformation{id=%s, imageID=%s, hash=%s, status=%s, spec=%s}",
		t.id, t.imageID, t.hash, t.status, specStr)
}
