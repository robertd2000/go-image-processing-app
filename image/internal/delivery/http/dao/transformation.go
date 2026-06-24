package dao

import (
	"encoding/json"
	"time"

	"github.com/robertd2000/go-image-processing-app/image/internal/usecase/transformation"
)

type TransformRequest struct {
	Spec json.RawMessage `json:"spec" binding:"required"`
}

type TransformResponse struct {
	TransformID string `json:"transform_id"`
	ImageID     string `json:"image_id"`
	Hash        string `json:"hash"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type TransformationStatusResponse struct {
	TransformID  string          `json:"transform_id"`
	ImageID      string          `json:"image_id"`
	Spec         json.RawMessage `json:"spec"`
	Hash         string          `json:"hash"`
	Status       string          `json:"status"`
	ResultKey    string          `json:"result_key,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	StartedAt    *string         `json:"started_at,omitempty"`
	CompletedAt  *string         `json:"completed_at,omitempty"`
	Duration     int64           `json:"duration,omitempty"`
	CreatedAt    string          `json:"created_at"`
}

func formatTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func ToTransformResponse(r *transformation.TransformationResult) TransformResponse {
	return TransformResponse{
		TransformID: r.ID.String(),
		ImageID:     r.ImageID.String(),
		Hash:        r.Hash,
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
	}
}

func ToTransformationStatusResponse(r *transformation.TransformationResult) TransformationStatusResponse {
	return TransformationStatusResponse{
		TransformID:  r.ID.String(),
		ImageID:      r.ImageID.String(),
		Spec:         r.Spec,
		Hash:         r.Hash,
		Status:       string(r.Status),
		ResultKey:    r.ResultKey,
		ErrorMessage: r.ErrorMessage,
		StartedAt:    formatTimePtr(r.StartedAt),
		CompletedAt:  formatTimePtr(r.CompletedAt),
		Duration:     r.Duration,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}
}
