package handler

import (
	"net/http"

	"github.com/NormaTech-AI/audity/services/client-service/internal/config"
	"github.com/NormaTech-AI/audity/services/client-service/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
    store  *store.Store
    config *config.Config
    logger *zap.SugaredLogger
    minio  *minio.Client
}

// NewHandler creates a new Handler instance
func NewHandler(store *store.Store, cfg *config.Config, logger *zap.SugaredLogger, minioClient *minio.Client) *Handler {
    return &Handler{
        store:  store,
        config: cfg,
        logger: logger,
        minio:  minioClient,
    }
}

// getTenantBaseURL returns the configured tenant service base URL
func (h *Handler) getTenantBaseURL() string {
    if h.config.Services.TenantBaseURL != "" {
        return h.config.Services.TenantBaseURL
    }
    // Fallback to docker-compose service name/port
    return "http://tenant-service:8081"
}

// RootHandler handles the root endpoint
func (h *Handler) RootHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"service": "client-service",
		"version": "1.0.0",
		"status":  "running",
	})
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c echo.Context) error {
	// Check database connection
	if err := h.store.Ping(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":   "healthy",
		"database": "connected",
	})
}
