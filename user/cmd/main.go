package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/robertd2000/go-image-processing-app/user/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robertd2000/go-image-processing-app/user/internal/config"
	deliveryHttp "github.com/robertd2000/go-image-processing-app/user/internal/delivery/http"
	v1 "github.com/robertd2000/go-image-processing-app/user/internal/delivery/http/v1"
	kafkahandler "github.com/robertd2000/go-image-processing-app/user/internal/delivery/kafka"
	ckafka "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/kafka"
	userpg "github.com/robertd2000/go-image-processing-app/user/internal/infrastructure/persistence/postgres/user"
	"github.com/robertd2000/go-image-processing-app/user/internal/usecase/user"
	"go.uber.org/zap"
)

// @title User Service API
// @version 1.0
// @description API for user service
// @host localhost:8083
// @BasePath /api/v1
func main() {
	// ---------- logger ----------
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

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

	// ---------- repos ----------
	userRepo := userpg.NewUserRepository(db)

	// ---------- service ----------
	userService := user.NewUserService(userRepo)

	// ---------- kafka ----------
	consumer := ckafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topics.UserCreated,
		cfg.Kafka.GroupID,
	)

	handler := kafkahandler.NewUserCreatedHandler(userService)
	userHandler := v1.NewUserHandler(userService, logger)

	deliveryHttp.SetupRouter(r, userHandler)

	go func() {
		logger.Info("Kafka consumer started")
		consumer.Start(appCtx, handler.Handle)
		logger.Info("Kafka consumer stopped")
	}()

	// ---------- http server ----------
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

	if err := consumer.Close(); err != nil {
		logger.Error("consumer close failed", zap.Error(err))
	}

	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	logger.Info("server exited properly")
}
