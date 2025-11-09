package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NormaTech-AI/audity/services/client-service/internal/config"
	"github.com/NormaTech-AI/audity/services/client-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/client-service/internal/router"
	"github.com/NormaTech-AI/audity/services/client-service/internal/store"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	var logger *zap.Logger
	if cfg.Logging.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize database connection pool
	dbConfig, err := pgxpool.ParseConfig(cfg.Database.GetDSN())
	if err != nil {
		sugar.Fatalw("Failed to parse database config", "error", err)
	}

	dbConfig.MaxConns = int32(cfg.Database.MaxConnections)
	dbConfig.MinConns = int32(cfg.Database.MinConnections)

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		sugar.Fatalw("Failed to create connection pool", "error", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		sugar.Fatalw("Failed to ping database", "error", err)
	}
	sugar.Info("Database connection established")

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		sugar.Fatalw("Failed to initialize MinIO client", "error", err)
	}
	sugar.Info("MinIO client initialized")

	// Initialize store
	st := store.NewStore(pool)

	// Initialize handler
	h := handler.NewHandler(st, cfg, sugar, minioClient)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = &CustomValidator{validator: validator.New()}

	// Setup routes
	router.SetupRoutes(e, h, cfg, st, sugar)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	sugar.Infow("Starting client-service", "address", serverAddr)

	// Graceful shutdown
	go func() {
		if err := e.Start(serverAddr); err != nil {
			sugar.Infow("Server stopped", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		sugar.Fatalw("Server forced to shutdown", "error", err)
	}

	sugar.Info("Server exited")
}
