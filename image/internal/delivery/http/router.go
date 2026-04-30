package http

import (
	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/middleware"
	v1 "github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/v1"
	"github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/auth"
)

func SetupRouter(r *gin.Engine, imageHandler *v1.ImageHandler, jwtValidator *auth.JWTValidator) {
	r.Use(gin.Recovery())
	authMiddleware := middleware.AuthMiddleware(jwtValidator)

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			imageHandler.SetupImageHandler(v1, authMiddleware)
		}
	}
}
