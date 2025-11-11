package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientstore"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/crypto"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/framework"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/migrations"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/router"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/store"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/validator"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/mail"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	_ "github.com/NormaTech-AI/audity/services/tenant-service/docs"
)

// @title Tenant Service API
// @version 1.0
// @description TPRM Audit Platform - Tenant Management Service
// @host localhost:8081
// @BasePath /
func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize zap logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	log := logger.Sugar()

	log.Info("Starting Tenant Service...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalw("Failed to load config", "error", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalw("Invalid configuration", "error", err)
	}

	log.Infow("Configuration loaded", "env", cfg.Server.Env)

	// Initialize Microsoft Graph API client
	mail.NewMailService(cfg, log)

	// Initialize database connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.TenantDBURL)
	if err != nil {
		log.Fatalw("Failed to parse database URL", "error", err)
	}

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalw("Failed to create connection pool", "error", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalw("Failed to ping database", "error", err)
	}
	log.Info("Successfully connected to tenant_db")

	// Initialize store
	st := store.NewStore(pool)

	// Initialize encryptor
	encryptor, err := crypto.NewEncryptor(cfg.Crypto.EncryptionKey)
	if err != nil {
		log.Fatalw("Failed to initialize encryptor", "error", err)
	}

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalw("Failed to initialize MinIO client", "error", err)
	}
	log.Info("Successfully connected to MinIO")

	// Initialize client migration runner
	clientMigrationRunner := migrations.NewClientMigrationRunner("./db/client-migrations", log)
	log.Info("Client migration runner initialized")

	// Initialize tenant DB queries for clientStore
	tenantQueries := db.New(pool)

	// Initialize client store
	clientStore := clientstore.NewClientStore(tenantQueries, encryptor, log)
	log.Info("Client store initialized")

	// Initialize framework service
	frameworkService := framework.NewService("./templates/frameworks", log)
	log.Info("Framework service initialized")

	// Initialize handler
	h := handler.NewHandler(st, cfg, encryptor, minioClient, log, clientMigrationRunner, clientStore, frameworkService)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.NewValidator()

	// Setup routes
	router.SetupRoutes(e, h, cfg, st, log)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Infow("Starting server", "address", addr)
		if err := e.Start(addr); err != nil {
			log.Infow("Server stopped", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited")
}