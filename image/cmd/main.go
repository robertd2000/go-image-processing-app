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
	"github.com/robertd2000/go-image-processing-app/image/internal/delivery/http/health"
	"github.com/segmentio/kafka-go"
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
	transformpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/transformation"
	txmanagerpg "github.com/robertd2000/go-image-processing-app/image/internal/infrastructure/persistence/postgtres/txmanager"
	imageSvc "github.com/robertd2000/go-image-processing-app/image/internal/usecase/image"
	transformSvc "github.com/robertd2000/go-image-processing-app/image/internal/usecase/transformation"
	"github.com/robertd2000/go-image-processing-app/image/internal/pkg/app"
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

	lc := app.NewLifecycle()

	// ---------- DB ----------
	db, err := pgxpool.New(appCtx, cfg.Postgres.DSN())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	lc.Add(app.CloserFunc(func(_ context.Context) error {
		db.Close()
		return nil
	}))

	if err := db.Ping(appCtx); err != nil {
		logger.Fatal("db ping failed", zap.Error(err))
	}

	if cfg.Server.RunMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
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
		publisher  port.EventPublisher
		consumer   port.EventConsumer
		outboxRepo port.OutboxRepository
		svcOpts    []imageSvc.ServiceOption
	)

	if cfg.Kafka.Enabled {
		outboxRepo = outboxpg.NewOutboxRepository(db, zlog)
		jobRepo := jobpg.NewJobRepository(db, zlog)

		svcOpts = append(svcOpts,
			imageSvc.WithOutbox(outboxRepo),
			imageSvc.WithJobRepo(jobRepo),
		)

		pub := kafkaAdapter.NewPublisher(cfg.Kafka.Brokers)
		con := kafkaAdapter.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID)
		publisher = pub
		consumer = con

		lc.Add(app.CloserFunc(func(_ context.Context) error { return pub.Close() }))
		lc.Add(app.CloserFunc(func(_ context.Context) error { return con.Close() }))
	}

	svc := imageSvc.NewImageService(imageRepo, st, metaExtractor, txManager, svcOpts...)

	if cfg.Kafka.Enabled {
		lc.Go(func(ctx context.Context) {
			infraEvents.RunOutboxRelay(
				ctx, outboxRepo, publisher,
				cfg.Kafka.Topics.ImageProcessingRequested,
				1*time.Second, 50,
			)
		}, appCtx)

		lc.Go(func(ctx context.Context) {
			handler := imageSvc.NewProcessingResultHandler(svc)
			handler = infraEvents.WithDLQ(handler, publisher, cfg.Kafka.Topics.ImageProcessedDLQ())
			if err := consumer.Consume(ctx, []string{cfg.Kafka.Topics.ImageProcessed}, handler); err != nil && err != context.Canceled {
				logger.Error("consumer stopped", zap.Error(err))
			}
		}, appCtx)
	}

	// ---------- health check ----------
	healthChecks := map[string]health.Check{
		"postgres": func(ctx context.Context) error { return db.Ping(ctx) },
	}
	if cfg.Kafka.Enabled {
		healthChecks["kafka"] = func(ctx context.Context) error {
			conn, err := kafka.DialContext(ctx, "tcp", cfg.Kafka.Brokers[0])
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		}
	}
	r.GET("/health", health.Handler(5*time.Second, healthChecks))

	// ---------- transformation service ----------
	transformRepo := transformpg.NewTransformationRepo(db, zlog)
	tSvc := transformSvc.NewService(imageRepo, transformRepo, txManager, outboxRepo)
	transformHandler := v1.NewTransformationHandler(tSvc, logger)

	imageHandler := v1.NewImageHandler(svc, logger)
	httpDelivery.SetupRouter(r, imageHandler, transformHandler, jwtValidator, &httpDelivery.RouterConfig{
		CORSOrigins:    []string{"*"},
		RequestTimeout: 30 * time.Second,
		RateLimit:      100,
		RateInterval:   1 * time.Minute,
		Logger:         logger,
	})

	// ---------- HTTP server ----------
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	lc.Add(app.CloserFunc(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}))

	go func() {
		logger.Info("server started", zap.String("addr", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// ---------- graceful shutdown ----------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("shutting down", zap.String("signal", sig.String()))

	// 1. Cancel app context — signals goroutines (relay, consumer) to stop
	cancel()

	// 2. Handle second signal for force exit
	go func() {
		<-quit
		logger.Warn("second signal received, forcing shutdown")
		os.Exit(1)
	}()

	// 3. Graceful shutdown: wait goroutines → close resources in reverse order
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := lc.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown errors", zap.Error(err))
	}

	logger.Info("server exited properly")
}
