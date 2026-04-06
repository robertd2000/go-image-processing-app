package http

import (
	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/user/internal/delivery/http/middleware"
	v1 "github.com/robertd2000/go-image-processing-app/user/internal/delivery/http/v1"
	"github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/auth"
)

func SetupRouter(r *gin.Engine, userHandler *v1.UserHandler, jwtValidator *auth.JWTValidator) {
	r.Use(gin.Recovery())
	authMiddleware := middleware.AuthMiddleware(jwtValidator)

	api := r.Group("/api")
	api.Use(authMiddleware)

	{
		v1 := api.Group("/v1")
		{
			userHandler.SetupUserHandler(v1)
		}
	}
}
