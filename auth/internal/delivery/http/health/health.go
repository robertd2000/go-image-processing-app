package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CheckResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Response struct {
	Status   string                 `json:"status"`
	Checks   map[string]CheckResult `json:"checks"`
	Duration string                 `json:"duration"`
}

type Check func(ctx context.Context) error

func Handler(timeout time.Duration, checks map[string]Check) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		results := make(map[string]CheckResult, len(checks))
		var mu sync.Mutex
		var wg sync.WaitGroup

		for name, check := range checks {
			wg.Add(1)
			go func(name string, check Check) {
				defer wg.Done()
				r := CheckResult{Status: "ok"}
				if err := check(ctx); err != nil {
					r.Status = "down"
					r.Error = err.Error()
				}
				mu.Lock()
				results[name] = r
				mu.Unlock()
			}(name, check)
		}

		wg.Wait()

		overall := "ok"
		for _, r := range results {
			if r.Status == "down" {
				overall = "degraded"
				break
			}
		}

		c.JSON(http.StatusOK, Response{
			Status:   overall,
			Checks:   results,
			Duration: time.Since(start).Round(time.Millisecond).String(),
		})
	}
}
