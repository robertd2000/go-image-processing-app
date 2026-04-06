package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/auth"
)

func AuthMiddleware(jwt *auth.JWTValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if header == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}

		header = strings.TrimSpace(header)

		var token string

		if strings.HasPrefix(strings.ToLower(header), "bearer ") {
			token = strings.TrimSpace(header[7:])
		} else {
			token = header
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		userID, err := jwt.ValidateAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
