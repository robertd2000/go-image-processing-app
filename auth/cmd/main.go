package main

import (
	"context"
	"fmt"
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
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/jwt"
	ekafka "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/kafka"
	outboxpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/outbox"
	tokenpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/token"
	postgres "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/txmanager"
	userpg "github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/persistence/postgres/user"
	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
	"github.com/robertd2000/go-image-processing-app/auth/internal/outbox"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/user"
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
	defer logger.Sync()

	// ---------- config ----------
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}

	// ---------- db ----------
	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	defer db.Close()

	// ---------- gin ----------
	if cfg.Server.RunMode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// ---------- swagger ----------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Kafka init
	broker := "kafka:9092"

	err = waitForKafka(broker, 10, 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	err = ekafka.EnsureTopic(broker, "user.created.v1")
	if err != nil {
		log.Fatal(err)
	}

	publisher := ekafka.NewKafkaPublisher([]string{broker})

	// repos
	userRepo := userpg.NewUserRepository(db)
	tokenRepo := tokenpg.NewTokenRepository(db)
	outboxRepo := outboxpg.NewRepository(db)

	// utils
	tokenGen := jwt.NewJWTGenerator([]byte(cfg.JWT.Secret))
	hasher := security.NewHasher()
	tokenHasher := &security.TokenHasher{}

	txManager := postgres.NewTxManager(db)

	// service
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
	userSvc := user.NewUserSyncService(userRepo)

	// outbox worker
	worker := outbox.NewWorker(outboxRepo, publisher)
	go worker.Start(ctx)

	consumer := ekafka.NewConsumer(
		[]string{"kafka:9092"},
		"auth-service",
		"user.status.updated.v1",
	)

	dispatcher := kafkahandler.NewDispatcher()

	dispatcher.Register("user.status.updated", kafkahandler.NewUserStatusChangeHandler(userSvc))

	go func() {
		err := consumer.Start(ctx, dispatcher.Dispatch)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// handler
	authHandler := v1.NewAuthHandler(authSvc, logger)
	// ---------- routes ----------
	delivery.SetupRouter(r, authHandler)

	// ---------- server ----------
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
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

	logger.Info("shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	logger.Info("server exited properly")
}

func waitForKafka(broker string, retries int, delay time.Duration) error {
	for range retries {
		conn, err := kafka.Dial("tcp", broker)
		if err == nil {
			conn.Close()
			log.Println("Kafka is ready")
			return nil
		}

		log.Println("waiting for Kafka...", err)
		time.Sleep(delay)
	}

	return fmt.Errorf("kafka is not available after retries")
}
