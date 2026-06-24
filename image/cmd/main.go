package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/image/internal/config"
	httpDelivery "github.com/robertd2000/go-image-processing-app/image/internal/delivery/http"
	v1 "github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/v1"
	imageinfra "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/auth"
	s3store "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/s3"
	imagepg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/image"
	txmanagerpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/txmanager"
	imageSvc "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Image Service API
// @version 1.0
// @description API for image service
// @host localhost:8081
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// ---------- logger ----------
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	zlog := logger.Sugar()

	// ---------- config ----------
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}

	// ---------- app context ----------
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---------- db ----------
	db, err := pgxpool.New(appCtx, cfg.Postgres.DSN())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(appCtx); err != nil {
		logger.Fatal("db ping failed", zap.Error(err))
	}

	// ---------- gin ----------
	if cfg.Server.RunMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	// ---------- swagger ----------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ---------- infrastructure ----------
	imageRepo := imagepg.NewImageRepository(db, zlog, nil)

	jwtValidator := auth.NewJWTValidator(cfg.JWT.Secret)

	s3Client := s3.NewFromConfig(aws.Config{
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if cfg.Storage.Endpoint != "" {
					return aws.Endpoint{
						URL:               cfg.Storage.Endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		),
	})

	st := s3store.New(s3Client, cfg.Storage.Bucket, cfg.Storage.Endpoint)

	metaExtractor := imageinfra.NewMetadataExtractor()

	txManager := txmanagerpg.NewTxManager(db, zlog)

	// ---------- usecase ----------
	svc := imageSvc.NewImageService(imageRepo, st, metaExtractor, txManager)

	// ---------- delivery ----------
	imageHandler := v1.NewImageHandler(svc, logger)

	httpDelivery.SetupRouter(r, imageHandler, jwtValidator)

	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", zap.String("addr", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// ---------- graceful shutdown ----------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("shutting down...")

	cancel()

	// if err := consumer.Close(); err != nil {
	// 	logger.Error("consumer close failed", zap.Error(err))
	// }

	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	logger.Info("server exited properly")
}
