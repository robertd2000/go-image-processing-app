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

		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func RequireRole(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesRaw, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "no roles"})
			return
		}

		roles, ok := rolesRaw.([]string)
		if !ok {
			c.AbortWithStatusJSON(500, gin.H{"error": "invalid roles type"})
			return
		}

		for _, r := range roles {
			if r == required {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
