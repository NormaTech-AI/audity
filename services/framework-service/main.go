package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NormaTech-AI/audity/services/framework-service/internal/config"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/router"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/store"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/validator"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	_ "github.com/NormaTech-AI/audity/services/framework-service/docs"
)

// @title Framework Service API
// @version 1.0
// @description TPRM Audit Platform - Framework Service
// @host localhost:8084
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

	log.Info("Starting Framework Service...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalw("Failed to load config", "error", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalw("Invalid configuration", "error", err)
	}

	// Run database migrations
	log.Info("Running database migrations...")
	dbURL := cfg.GetDatabaseURL()
	m, err := migrate.New(
		"file://db/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalw("Failed to create migrate instance", "error", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalw("Failed to run migrations", "error", err)
	}
	log.Info("Migrations completed successfully")

	// Initialize database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalw("Failed to ping database", "error", err)
	}
	log.Info("Database connection established")

	// Initialize store
	st := store.NewStore(pool)

	// Initialize handler
	h := handler.NewHandler(st, cfg, log)

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
