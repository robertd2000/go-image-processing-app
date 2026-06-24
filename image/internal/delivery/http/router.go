package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/middleware"
	v1 "github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/v1"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	"go.uber.org/zap"
)

type RouterConfig struct {
	CORSOrigins    []string
	RequestTimeout time.Duration
	RateLimit      int
	RateInterval   time.Duration
	Logger         *zap.Logger
}

func SetupRouter(r *gin.Engine, imageHandler *v1.ImageHandler, tokenValidator port.TokenValidator, cfg *RouterConfig) {
	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(middleware.Logger(cfg.Logger))
	r.Use(gin.Recovery())

	authMiddleware := middleware.AuthMiddleware(tokenValidator)

	api := r.Group("/api")
	api.Use(middleware.RateLimit(cfg.RateLimit, cfg.RateInterval))

	{
		v1 := api.Group("/v1")
		{
			imageHandler.SetupImageHandler(v1, authMiddleware)
		}
	}
}
