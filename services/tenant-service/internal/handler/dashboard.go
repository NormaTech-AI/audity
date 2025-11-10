package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalFrameworks       int64 `json:"total_frameworks"`
	TotalClients          int64 `json:"total_clients"`
	TotalUsers            int64 `json:"total_users"`
	TotalClientFrameworks int64 `json:"total_client_frameworks"`
	TotalAuditLogs        int64 `json:"total_audit_logs"`
}

// DashboardData represents complete dashboard data
type DashboardData struct {
	Stats            DashboardStats           `json:"stats"`
	RecentActivities []map[string]interface{} `json:"recent_activities"`
}

// GetTenantDashboard godoc
// @Summary Get tenant dashboard data
// @Description Get dashboard statistics and recent activities for tenant service
// @Tags dashboard
// @Produce json
// @Success 200 {object} DashboardData
// @Failure 500 {object} map[string]string
// @Router /api/tenant/dashboard [get]
func (h *Handler) GetTenantDashboard(c echo.Context) error {
	ctx := c.Request().Context()

	// Framework count will be fetched from framework-service if needed
	frameworkCount := int64(0)

	// Get client count
	clientCount, err := h.store.CountClients(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count clients", "error", err)
		clientCount = 0
	}

	// Get user count
	userCount, err := h.store.CountTotalUsers(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count users", "error", err)
		userCount = 0
	}

	// Get client frameworks count
	clientFrameworkCount, err := h.store.CountClientFrameworks(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client frameworks", "error", err)
		clientFrameworkCount = 0
	}

	// Get audit logs count
	auditLogCount, err := h.store.CountAuditLogs(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count audit logs", "error", err)
		auditLogCount = 0
	}

	stats := DashboardStats{
		TotalFrameworks:       frameworkCount,
		TotalClients:          clientCount,
		TotalUsers:            userCount,
		TotalClientFrameworks: clientFrameworkCount,
		TotalAuditLogs:        auditLogCount,
	}

	// TODO: Fetch recent activities from activity log
	recentActivities := []map[string]interface{}{}

	dashboardData := DashboardData{
		Stats:            stats,
		RecentActivities: recentActivities,
	}

	return c.JSON(http.StatusOK, dashboardData)
}

// GetTenantDashboardStats godoc
// @Summary Get tenant dashboard statistics
// @Description Get only statistics for tenant service dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} DashboardStats
// @Failure 500 {object} map[string]string
// @Router /api/tenant/dashboard/stats [get]
func (h *Handler) GetTenantDashboardStats(c echo.Context) error {
	ctx := c.Request().Context()

	// Framework count will be fetched from framework-service if needed
	frameworkCount := int64(0)

	// Get client count
	clientCount, err := h.store.CountClients(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count clients", "error", err)
		clientCount = 0
	}

	// Get user count
	userCount, err := h.store.CountTotalUsers(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count users", "error", err)
		userCount = 0
	}

	// Get client frameworks count
	clientFrameworkCount, err := h.store.CountClientFrameworks(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count client frameworks", "error", err)
		clientFrameworkCount = 0
	}

	// Get audit logs count
	auditLogCount, err := h.store.CountAuditLogs(ctx)
	if err != nil {
		h.logger.Errorw("Failed to count audit logs", "error", err)
		auditLogCount = 0
	}

	stats := DashboardStats{
		TotalFrameworks:       frameworkCount,
		TotalClients:          clientCount,
		TotalUsers:            userCount,
		TotalClientFrameworks: clientFrameworkCount,
		TotalAuditLogs:        auditLogCount,
	}

	return c.JSON(http.StatusOK, stats)
}
