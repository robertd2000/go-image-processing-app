package delivery

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/http/v1"
)

func SetupRouter(r *gin.Engine, authHandler *v1.AuthHandler) {
	r.Use(gin.Recovery())

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			authHandler.SetupAuthHandler(v1)
		}
	}
}
