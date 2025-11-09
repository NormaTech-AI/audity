package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/config"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/oidc"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/router"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/store"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	_ "github.com/NormaTech-AI/audity/services/auth-service/docs"
)

// @title Auth Service API
// @version 1.0
// @description TPRM Audit Platform - Authentication Service
// @host localhost:8082
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

	log.Info("Starting Auth Service...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalw("Failed to load config", "error", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalw("Invalid configuration", "error", err)
	}

	log.Infow("Configuration loaded", "env", cfg.Server.Env)

	// Initialize database connection pool
	log.Info("Connecting to tenant_db...")
	log.Infof("Database URL: %s", cfg.Database.TenantDBURL)
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

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpirationHours)

	// Initialize OIDC providers
	var googleProvider, microsoftProvider *oidc.OIDCProvider

	if cfg.Auth.GoogleClientID != "" && cfg.Auth.GoogleClientSecret != "" {
		googleProvider = oidc.NewGoogleProvider(
			cfg.Auth.GoogleClientID,
			cfg.Auth.GoogleClientSecret,
			cfg.Auth.RedirectURL,
		)
		log.Info("Google OAuth provider initialized")
	} else {
		log.Warn("Google OAuth not configured")
	}

	if cfg.Auth.MicrosoftClientID != "" && cfg.Auth.MicrosoftClientSecret != "" {
		microsoftProvider = oidc.NewMicrosoftProvider(
			cfg.Auth.MicrosoftClientID,
			cfg.Auth.MicrosoftClientSecret,
			cfg.Auth.RedirectURL,
		)
		log.Info("Microsoft OAuth provider initialized")
	} else {
		log.Warn("Microsoft OAuth not configured")
	}

	// Initialize handler
	h := handler.NewHandler(st, cfg, jwtManager, googleProvider, microsoftProvider, log)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.NewValidator()

	// Setup routes
	router.SetupRoutes(e, h, cfg.Auth.JWTSecret, log)

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
