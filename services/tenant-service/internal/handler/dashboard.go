package handler

import (
	"net/http"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/google/uuid"
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

// Client-specific dashboard types

// AuditCycleEnrollment represents a client's enrollment in an audit cycle
type AuditCycleEnrollment struct {
	AuditCycleID          uuid.UUID  `json:"audit_cycle_id"`
	AuditCycleName        string     `json:"audit_cycle_name"`
	AuditCycleDescription *string    `json:"audit_cycle_description"`
	StartDate             time.Time  `json:"start_date"`
	EndDate               time.Time  `json:"end_date"`
	CycleStatus           string     `json:"cycle_status"`
	EnrollmentID          uuid.UUID  `json:"enrollment_id"`
	EnrolledAt            time.Time  `json:"enrolled_at"`
	Frameworks            []FrameworkAssignment `json:"frameworks"`
}

// FrameworkAssignment represents a framework assigned to a client in an audit cycle
type FrameworkAssignment struct {
	FrameworkAssignmentID *uuid.UUID `json:"framework_assignment_id,omitempty"`
	FrameworkID           *uuid.UUID `json:"framework_id,omitempty"`
	FrameworkName         *string    `json:"framework_name,omitempty"`
	DueDate               *time.Time `json:"due_date,omitempty"`
	FrameworkStatus       *string    `json:"framework_status,omitempty"`
	AuditorID             *uuid.UUID `json:"auditor_id,omitempty"`
}

// FrameworkAnalytics represents analytics for a framework
type FrameworkAnalytics struct {
	AuditID           uuid.UUID  `json:"audit_id"`
	FrameworkID       uuid.UUID  `json:"framework_id"`
	FrameworkName     string     `json:"framework_name"`
	Status            string     `json:"status"`
	DueDate           time.Time  `json:"due_date"`
	TotalQuestions    int64      `json:"total_questions"`
	AnsweredQuestions int64      `json:"answered_questions"`
}

// ClientDashboardStats represents dashboard statistics for a client
type ClientDashboardStats struct {
	ActiveAuditCycles         int64 `json:"active_audit_cycles"`
	TotalFrameworkAssignments int64 `json:"total_framework_assignments"`
}

// ClientDashboardData represents complete dashboard data for a client
type ClientDashboardData struct {
	ClientName       string                 `json:"client_name"`
	Stats            ClientDashboardStats   `json:"stats"`
	AuditCycles      []AuditCycleEnrollment `json:"audit_cycles"`
	FrameworkAnalytics []FrameworkAnalytics `json:"framework_analytics"`
}

// GetClientDashboard godoc
// @Summary Get client-specific dashboard data
// @Description Get dashboard data for a specific client including audit cycle enrollments and framework analytics
// @Tags dashboard
// @Produce json
// @Param client_id path string true "Client ID"
// @Success 200 {object} ClientDashboardData
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tenant/dashboard/client/{client_id} [get]
func (h *Handler) GetClientDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Get client ID from path parameter
	clientIDStr := c.Param("client_id")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid client ID"})
	}

	// Get client information
	client, err := h.store.GetClient(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client", "error", err, "client_id", clientID)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
	}

	// Get active audit cycles count
	activeCount, err := h.store.CountClientActiveAuditCycles(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to count active audit cycles", "error", err, "client_id", clientID)
		activeCount = 0
	}

	// Get total framework assignments count
	frameworkCount, err := h.store.CountClientTotalFrameworkAssignments(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to count framework assignments", "error", err, "client_id", clientID)
		frameworkCount = 0
	}

	stats := ClientDashboardStats{
		ActiveAuditCycles:         activeCount,
		TotalFrameworkAssignments: frameworkCount,
	}

	// Get audit cycle enrollments
	enrollments, err := h.store.GetClientAuditCycleEnrollments(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get audit cycle enrollments", "error", err, "client_id", clientID)
		enrollments = []db.GetClientAuditCycleEnrollmentsRow{}
	}

	// Group enrollments by audit cycle
	auditCycleMap := make(map[uuid.UUID]*AuditCycleEnrollment)
	for _, enrollment := range enrollments {
		cycleID := enrollment.AuditCycleID
		
		// Create or get audit cycle entry
		if _, exists := auditCycleMap[cycleID]; !exists {
			startTime, _ := enrollment.StartDate.Value()
			endTime, _ := enrollment.EndDate.Value()
			enrolledTime, _ := enrollment.EnrolledAt.Value()
			
			auditCycleMap[cycleID] = &AuditCycleEnrollment{
				AuditCycleID:          enrollment.AuditCycleID,
				AuditCycleName:        enrollment.AuditCycleName,
				AuditCycleDescription: enrollment.AuditCycleDescription,
				StartDate:             startTime.(time.Time),
				EndDate:               endTime.(time.Time),
				CycleStatus:           *enrollment.CycleStatus,
				EnrollmentID:          enrollment.EnrollmentID,
				EnrolledAt:            enrolledTime.(time.Time),
				Frameworks:            []FrameworkAssignment{},
			}
		}
		
		// Add framework if present
		if enrollment.FrameworkAssignmentID.Valid {
			var dueDate *time.Time
			if enrollment.DueDate.Valid {
				dueDateVal, _ := enrollment.DueDate.Value()
				dt := dueDateVal.(time.Time)
				dueDate = &dt
			}
			
			var frameworkID *uuid.UUID
			if enrollment.FrameworkID.Valid {
				fid := enrollment.FrameworkID.Bytes
				uid, _ := uuid.FromBytes(fid[:])
				frameworkID = &uid
			}
			
			var auditorID *uuid.UUID
			if enrollment.AuditorID.Valid {
				aid := enrollment.AuditorID.Bytes
				uid, _ := uuid.FromBytes(aid[:])
				auditorID = &uid
			}
			
			faid := enrollment.FrameworkAssignmentID.Bytes
			faUUID, _ := uuid.FromBytes(faid[:])
			
			framework := FrameworkAssignment{
				FrameworkAssignmentID: &faUUID,
				FrameworkID:           frameworkID,
				FrameworkName:         enrollment.FrameworkName,
				DueDate:               dueDate,
				FrameworkStatus:       enrollment.FrameworkStatus,
				AuditorID:             auditorID,
			}
			
			auditCycleMap[cycleID].Frameworks = append(auditCycleMap[cycleID].Frameworks, framework)
		}
	}

	auditCycles := make([]AuditCycleEnrollment, 0, len(auditCycleMap))
	for _, cycle := range auditCycleMap {
		auditCycles = append(auditCycles, *cycle)
	}

	// Get framework analytics from client database
	frameworkAnalytics := []FrameworkAnalytics{}
	
	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
	} else {
		// Query framework analytics
		analyticsRows, err := clientQueries.GetAllAuditsProgress(ctx)
		if err != nil {
			h.logger.Errorw("Failed to get framework analytics", "error", err, "client_id", clientID)
		} else {
			for _, row := range analyticsRows {
				dueDate, _ := row.DueDate.Value()
				frameworkAnalytics = append(frameworkAnalytics, FrameworkAnalytics{
					AuditID:           row.ID,
					FrameworkID:       row.FrameworkID,
					FrameworkName:     row.FrameworkName,
					Status:            string(row.Status),
					DueDate:           dueDate.(time.Time),
					TotalQuestions:    row.TotalQuestions,
					AnsweredQuestions: row.AnsweredQuestions,
				})
			}
		}
	}

	dashboardData := ClientDashboardData{
		ClientName:         client.Name,
		Stats:              stats,
		AuditCycles:        auditCycles,
		FrameworkAnalytics: frameworkAnalytics,
	}

	return c.JSON(http.StatusOK, dashboardData)
}
