package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	roleDomain "github.com/robertd2000/go-image-processing-app/user/internal/domain/role"
	"github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/auth"
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

func RequireRole(required roleDomain.Name) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesRaw, exists := c.Get(string(ContextRoles))
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
			if roleDomain.Name(r) == required {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}

func extractRoles(c *gin.Context) ([]roleDomain.Role, error) {
	raw, exists := c.Get(string(ContextRoles))
	if !exists {
		return nil, errors.New("roles not found in context")
	}

	names, ok := raw.([]string)
	if !ok {
		return nil, errors.New("invalid roles type")
	}

	roles := make([]roleDomain.Role, 0, len(names))

	for _, name := range names {
		r, err := roleDomain.FromName(name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	return roles, nil
}

func RequirePermission(p roleDomain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := extractRoles(c)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		for _, r := range roles {
			if r.HasPermission(p) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}

func OwnerOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet(string(ContextUserID)).(uuid.UUID)
		roles := c.MustGet(string(ContextRoles)).([]string)

		targetID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": "invalid id"})
			return
		}

		// owner
		if userID == targetID {
			c.Next()
			return
		}

		// admin
		for _, r := range roles {
			if roleDomain.Name(r) == roleDomain.Admin {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(string(ContextRoles))
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		roles, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(500, gin.H{"error": "invalid roles type"})
			return
		}

		for _, r := range roles {
			if roleDomain.Name(r) == roleDomain.Admin {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
