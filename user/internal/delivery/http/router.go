package http

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/robertd2000/go-image-processing-app/user/internal/delivery/http/v1"
)

func SetupRouter(r *gin.Engine, userHandler *v1.UserHandler) {
	r.Use(gin.Recovery())

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			userHandler.SetupUserHandler(v1)
		}
	}
}
