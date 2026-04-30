package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/auth"
)

type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextRoles  contextKey = "roles"
)

func AuthMiddleware(jwt *auth.JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if header == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		var token string
		if strings.HasPrefix(strings.ToLower(header), "bearer ") {
			token = strings.TrimSpace(header[7:])
		} else {
			token = strings.TrimSpace(header)
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		claims, err := jwt.ValidateAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextRoles), claims.Roles)

		c.Next()
	}
}
