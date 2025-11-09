package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DashboardStats represents dashboard statistics for client service
type DashboardStats struct {
	TotalClients int64 `json:"total_clients"`
	ActiveClients int64 `json:"active_clients"`
	InactiveClients int64 `json:"inactive_clients"`
	TotalClientDatabases int64 `json:"total_client_databases"`
	TotalClientBuckets int64 `json:"total_client_buckets"`
}

// DashboardData represents complete dashboard data
type DashboardData struct {
	Stats            DashboardStats           `json:"stats"`
	RecentActivities []map[string]interface{} `json:"recent_activities"`
}

// GetClientDashboard godoc
// @Summary Get client dashboard data
// @Description Get dashboard statistics and recent activities for client service
// @Tags dashboard
// @Produce json
// @Success 200 {object} DashboardData
// @Failure 500 {object} map[string]string
// @Router /api/client/dashboard [get]
func (h *Handler) GetClientDashboard(c echo.Context) error {
	ctx := c.Request().Context()

	// Get total client count
	clientCount, err := h.store.Queries().CountClients(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count clients", "error", err)
		clientCount = 0
	}

	// Get client database count
	clientDBCount, err := h.store.Queries().CountClientDatabases(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client databases", "error", err)
		clientDBCount = 0
	}

	// Get client bucket count
	clientBucketCount, err := h.store.Queries().CountClientBuckets(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client buckets", "error", err)
		clientBucketCount = 0
	}

	stats := DashboardStats{
		TotalClients:         clientCount,
		ActiveClients:        0, // TODO: Implement status-based counting
		InactiveClients:      0, // TODO: Implement status-based counting
		TotalClientDatabases: clientDBCount,
		TotalClientBuckets:   clientBucketCount,
	}

	// TODO: Fetch recent activities
	recentActivities := []map[string]interface{}{}

	dashboardData := DashboardData{
		Stats:            stats,
		RecentActivities: recentActivities,
	}

	return c.JSON(http.StatusOK, dashboardData)
}

// GetClientDashboardStats godoc
// @Summary Get client dashboard statistics
// @Description Get only statistics for client service dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} DashboardStats
// @Failure 500 {object} map[string]string
// @Router /api/client/dashboard/stats [get]
func (h *Handler) GetClientDashboardStats(c echo.Context) error {
	ctx := c.Request().Context()

	// Get total client count
	clientCount, err := h.store.Queries().CountClients(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count clients", "error", err)
		clientCount = 0
	}

	// Get client database count
	clientDBCount, err := h.store.Queries().CountClientDatabases(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client databases", "error", err)
		clientDBCount = 0
	}

	// Get client bucket count
	clientBucketCount, err := h.store.Queries().CountClientBuckets(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client buckets", "error", err)
		clientBucketCount = 0
	}

	stats := DashboardStats{
		TotalClients:         clientCount,
		ActiveClients:        0, // TODO: Implement status-based counting
		InactiveClients:      0, // TODO: Implement status-based counting
		TotalClientDatabases: clientDBCount,
		TotalClientBuckets:   clientBucketCount,
	}

	return c.JSON(http.StatusOK, stats)
}
