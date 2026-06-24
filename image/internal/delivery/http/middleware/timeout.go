package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() != nil {
			if !c.IsAborted() {
				c.AbortWithStatusJSON(408, gin.H{"error": "request timeout"})
			}
		}
	}
}
