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
	"github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/auth"
	infraEvents "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/events"
	kafkaAdapter "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/events/kafka"
	imageinfra "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/image"
	jobpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/job"
	outboxpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/outbox"
	s3store "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/s3"
	imagepg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/image"
	txmanagerpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/txmanager"
	imageSvc "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	zlog := logger.Sugar()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pgxpool.New(appCtx, cfg.Postgres.DSN())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(appCtx); err != nil {
		logger.Fatal("db ping failed", zap.Error(err))
	}

	if cfg.Server.RunMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
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

	// ---------- events (optional) ----------
	var (
		publisher port.EventPublisher
		consumer  port.EventConsumer
	)

	var svcOpts []imageSvc.ServiceOption

	if cfg.Kafka.Enabled {
		outboxRepo := outboxpg.NewOutboxRepository(db, zlog)
		jobRepo := jobpg.NewJobRepository(db, zlog)

		svcOpts = append(svcOpts,
			imageSvc.WithOutbox(outboxRepo),
			imageSvc.WithJobRepo(jobRepo),
		)

		pub := kafkaAdapter.NewPublisher(cfg.Kafka.Brokers)
		con := kafkaAdapter.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID)
		publisher = pub
		consumer = con

		// usecase
		svc := imageSvc.NewImageService(imageRepo, st, metaExtractor, txManager, svcOpts...)

		// outbox relay
		go infraEvents.RunOutboxRelay(
			appCtx, outboxRepo, pub,
			cfg.Kafka.Topics.ImageProcessingRequested,
			1*time.Second, 50,
		)

		// consumer: listening for processing results
		go func() {
			handler := imageSvc.NewProcessingResultHandler(svc)
			if err := con.Consume(appCtx, []string{cfg.Kafka.Topics.ImageProcessed}, handler); err != nil && err != context.Canceled {
				logger.Error("consumer stopped", zap.Error(err))
			}
		}()

		// delivery
		imageHandler := v1.NewImageHandler(svc, logger)
		httpDelivery.SetupRouter(r, imageHandler, jwtValidator)
	} else {
		svc := imageSvc.NewImageService(imageRepo, st, metaExtractor, txManager)
		imageHandler := v1.NewImageHandler(svc, logger)
		httpDelivery.SetupRouter(r, imageHandler, jwtValidator)
	}

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("shutting down...")
	cancel()

	if publisher != nil {
		_ = publisher.Close()
	}
	if consumer != nil {
		_ = consumer.Close()
	}

	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	logger.Info("server exited properly")
}
