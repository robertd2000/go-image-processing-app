package delivery

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/auth/internal/delivery/http/middleware"
	v1 "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/http/v1"
	"go.uber.org/zap"
)

type RouterConfig struct {
	RequestTimeout time.Duration
	Logger         *zap.Logger
}

func SetupRouter(r *gin.Engine, authHandler *v1.AuthHandler, cfg *RouterConfig) {
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(gin.Recovery())

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			authHandler.SetupAuthHandler(v1)
		}
	}
}
