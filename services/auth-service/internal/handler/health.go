package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Service  string `json:"service"`
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service and database are healthy
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 500 {object} map[string]string
// @Router /health [get]
func (h *Handler) HealthCheck(c echo.Context) error {
	// Check database connection
	if err := h.store.Ping(c.Request().Context()); err != nil {
		h.logger.Errorw("Database health check failed", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"status":   "error",
			"database": "disconnected",
			"service":  "auth-service",
		})
	}

	return c.JSON(http.StatusOK, HealthResponse{
		Status:   "ok",
		Database: "connected",
		Service:  "auth-service",
	})
}

// RootHandler godoc
// @Summary Root endpoint
// @Description Service information
// @Tags info
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func (h *Handler) RootHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"service": "auth-service",
		"status":  "running",
		"version": "1.0.0",
	})
}
