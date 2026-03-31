package delivery

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/auth/internal/config"
	v1 "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/v1"
	elog "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/events/log"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	tokenpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/token"
	userpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"go.uber.org/zap"
)

func SetupRouter(r *gin.Engine, cfg *config.Config, db *pgxpool.Pool, logger *zap.Logger) {
	r.Use(gin.Recovery())

	userRepo := userpg.NewUserRepository(db)
	tokenRepo := tokenpg.NewTokenRepository(db)

	tokenGen := jwt.NewJWTGenerator([]byte(cfg.JWT.Secret))
	hasher := security.NewHasher()
	tokenHasher := &security.TokenHasher{}
	eventPublisher := elog.NewPublisher()

	authSvc := auth.NewAuthService(userRepo, tokenRepo, hasher, tokenHasher, tokenGen, eventPublisher, time.Duration(cfg.JWT.AccessTTLMin)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLMin)*time.Minute)

	authHandler := v1.NewAuthHandler(authSvc, logger)

	api := r.Group("/api")

	{
		v1 := api.Group("/v1")
		{
			authHandler.SetupAuthHandler(v1)
		}
	}
}
