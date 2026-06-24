package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/robertd2000/go-image-processing-app/auth/docs"
	"github.com/segmentio/kafka-go"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/robertd2000/go-image-processing-app/auth/internal/config"
	"github.com/robertd2000/go-image-processing-app/auth/internal/delivery"
	v1 "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/http/v1"
	kafkahandler "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/kafka"
	kafkamiddleware "github.com/robertd2000/go-image-processing-app/auth/internal/delivery/kafka/middleware"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	ekafka "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/kafka"
	outboxpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/outbox"
	tokenpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/token"
	txmanagerpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/txmanager"
	userpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/outbox"
	"github.com/robertd2000/go-image-processing-app/auth/internal/pkg/app"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/delivery/http/health"
)

// @title Auth Service API
// @version 1.0
// @description Auth service for image processing app
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// ---------- logger ----------
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()
	zlog := logger.Sugar()

	// ---------- config ----------
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lc := app.NewLifecycle()

	// ---------- db ----------
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

	// ---------- gin ----------
	if cfg.Server.RunMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ---------- swagger ----------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ---------- Kafka init ----------
	broker := cfg.Kafka.Brokers[0]

	if err := waitForKafka(broker, 10, 2*time.Second); err != nil {
		logger.Fatal("kafka not ready", zap.Error(err))
	}

	if err := ekafka.EnsureTopic(broker, cfg.Kafka.Topics.UserEvents); err != nil {
		logger.Fatal("ensure topic failed", zap.Error(err))
	}

	if err := ekafka.EnsureTopic(broker, cfg.Kafka.Topics.UserEventsDLQ()); err != nil {
		logger.Fatal("ensure dlq topic failed", zap.Error(err))
	}

	publisher := ekafka.NewKafkaPublisher(cfg.Kafka.Brokers)
	lc.Add(app.CloserFunc(func(_ context.Context) error { return publisher.Close() }))

	// ---------- repos ----------
	userRepo := userpg.NewUserRepository(db, zlog)
	tokenRepo := tokenpg.NewTokenRepository(db, zlog)
	outboxRepo := outboxpg.NewRepository(db)

	// ---------- utils ----------
	tokenGen := jwt.NewJWTGenerator([]byte(cfg.JWT.Secret))
	hasher := security.NewHasher()
	tokenHasher := &security.TokenHasher{}
	txManager := txmanagerpg.NewTxManager(db, zlog)

	// ---------- services ----------
	authSvc := auth.NewAuthService(
		userRepo,
		tokenRepo,
		outboxRepo,
		hasher,
		tokenHasher,
		tokenGen,
		time.Duration(cfg.JWT.AccessTTLMin)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLMin)*time.Minute,
		txManager,
	)
	userSvc := user.NewUserSyncService(txManager, userRepo, tokenRepo)

	// ---------- outbox worker ----------
	worker := outbox.NewWorker(outboxRepo, publisher)
	lc.Go(worker.Start, appCtx)

	// ---------- kafka consumer ----------
	consumer := ekafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
		cfg.Kafka.Topics.UserEvents,
	)
	lc.Add(app.CloserFunc(func(_ context.Context) error {
		consumer.Close()
		return nil
	}))

	dispatcher := kafkahandler.NewDispatcher()

	dlq := kafkamiddleware.NewDLQProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topics.UserEventsDLQ(),
	)
	lc.Add(app.CloserFunc(func(_ context.Context) error { return dlq.Close() }))

	dispatcher.Use(kafkamiddleware.DLQMiddleware(dlq))

	dispatcher.Use(kafkamiddleware.RetryMiddleware(kafkamiddleware.RetryConfig{
		MaxAttempts: 3,
		Backoff:     500 * time.Millisecond,
	}))

	dispatcher.Register(
		"user.deleted",
		kafkahandler.NewUserDeletedHandler(userSvc),
	)
	dispatcher.Register(
		"user.banned",
		kafkahandler.NewUserBanHandler(userSvc),
	)
	dispatcher.Register(
		"user.restored",
		kafkahandler.NewUserRestoreHandler(userSvc),
	)
	dispatcher.Register(
		"user.unbanned",
		kafkahandler.NewUserUnbanHandler(userSvc),
	)

	lc.Go(func(ctx context.Context) {
		if err := consumer.Start(ctx, dispatcher.Dispatch); err != nil && err != context.Canceled {
			logger.Error("consumer stopped", zap.Error(err))
		}
	}, appCtx)

	// ---------- health check ----------
	healthChecks := map[string]health.Check{
		"postgres": func(ctx context.Context) error { return db.Ping(ctx) },
		"kafka": func(ctx context.Context) error {
			conn, err := kafka.DialContext(ctx, "tcp", cfg.Kafka.Brokers[0])
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		},
	}
	r.GET("/health", health.Handler(5*time.Second, healthChecks))

	// ---------- HTTP handler + router ----------
	authHandler := v1.NewAuthHandler(authSvc, logger)
	delivery.SetupRouter(r, authHandler, &delivery.RouterConfig{
		RequestTimeout: 30 * time.Second,
		Logger:         logger,
	})

	// ---------- HTTP server ----------
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
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

	cancel()

	go func() {
		<-quit
		logger.Warn("second signal received, forcing shutdown")
		os.Exit(1)
	}()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := lc.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown errors", zap.Error(err))
	}

	logger.Info("server exited properly")
}

func waitForKafka(broker string, retries int, delay time.Duration) error {
	for range retries {
		conn, err := kafka.Dial("tcp", broker)
		if err == nil {
			conn.Close()
			return nil
		}
		log.Println("waiting for Kafka...", err)
		time.Sleep(delay)
	}
	_, err := kafka.Dial("tcp", broker)
	return err
}
