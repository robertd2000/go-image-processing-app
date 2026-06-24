package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_AllOk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Handler(0, map[string]Check{
		"postgres": func(_ context.Context) error { return nil },
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "ok", resp.Checks["postgres"].Status)
	assert.Empty(t, resp.Checks["postgres"].Error)
	assert.NotEmpty(t, resp.Duration)
}

func TestHealthHandler_OneDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Handler(0, map[string]Check{
		"postgres": func(_ context.Context) error { return nil },
		"kafka":    func(_ context.Context) error { return errors.New("connection refused") },
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "ok", resp.Checks["postgres"].Status)
	assert.Equal(t, "down", resp.Checks["kafka"].Status)
	assert.Contains(t, resp.Checks["kafka"].Error, "connection refused")
}

func TestHealthHandler_Timeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Handler(10*time.Millisecond, map[string]Check{
		"slow": func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "degraded", resp.Status)
	assert.Equal(t, "down", resp.Checks["slow"].Status)
}
