package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// ============================================================================
// Request/Response Types
// ============================================================================

type CreateAuditCycleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date" validate:"required"`
}

type UpdateAuditCycleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Status      *string `json:"status"`
}

type AddClientToAuditCycleRequest struct {
	ClientID string `json:"client_id" validate:"required"`
}

type AssignFrameworkRequest struct {
	FrameworkID   string  `json:"framework_id" validate:"required"`
	FrameworkName string  `json:"framework_name" validate:"required"`
	DueDate       *string `json:"due_date"`
	AuditorID     *string `json:"auditor_id"`
}

type AuditCycleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Status      string    `json:"status"`
	CreatedBy   *string   `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuditCycleClientResponse struct {
	ID             string    `json:"id"`
	AuditCycleID   string    `json:"audit_cycle_id"`
	ClientID       string    `json:"client_id"`
	ClientName     string    `json:"client_name"`
	POCEmail       string    `json:"poc_email"`
	ClientStatus   string    `json:"client_status"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuditCycleFrameworkResponse struct {
	ID                  string     `json:"id"`
	AuditCycleClientID  string     `json:"audit_cycle_client_id"`
	FrameworkID         string     `json:"framework_id"`
	FrameworkName       string     `json:"framework_name"`
	ClientID            string     `json:"client_id"`
	ClientName          string     `json:"client_name"`
	AssignedBy          *string    `json:"assigned_by"`
	AssignedAt          time.Time  `json:"assigned_at"`
	DueDate             *string    `json:"due_date"`
	Status              string     `json:"status"`
	AuditorID           *string    `json:"auditor_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type AuditCycleStatsResponse struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Status                string `json:"status"`
	TotalClients          int64  `json:"total_clients"`
	TotalFrameworks       int64  `json:"total_frameworks"`
	CompletedFrameworks   int64  `json:"completed_frameworks"`
	InProgressFrameworks  int64  `json:"in_progress_frameworks"`
	PendingFrameworks     int64  `json:"pending_frameworks"`
	OverdueFrameworks     int64  `json:"overdue_frameworks"`
}

// ============================================================================
// Handlers
// ============================================================================

// CreateAuditCycle creates a new audit cycle
// @Summary Create audit cycle
// @Tags audit-cycles
// @Accept json
// @Produce json
// @Param request body CreateAuditCycleRequest true "Audit cycle details"
// @Success 201 {object} AuditCycleResponse
// @Router /api/audit-cycles [post]
func (h *Handler) CreateAuditCycle(c echo.Context) error {
	var req CreateAuditCycleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be after start_date")
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Create audit cycle
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	cycle, err := h.store.Queries.CreateAuditCycle(c.Request().Context(), db.CreateAuditCycleParams{
		Name:        req.Name,
		Description: description,
		StartDate:   pgtype.Date{Time: startDate, Valid: true},
		EndDate:     pgtype.Date{Time: endDate, Valid: true},
		CreatedBy:   pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		h.logger.Errorw("Failed to create audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create audit cycle")
	}

	return c.JSON(http.StatusCreated, convertToAuditCycleResponse(cycle))
}

// ListAuditCycles lists all audit cycles
// @Summary List audit cycles
// @Tags audit-cycles
// @Produce json
// @Success 200 {array} AuditCycleResponse
// @Router /api/audit-cycles [get]
func (h *Handler) ListAuditCycles(c echo.Context) error {
	cycles, err := h.store.Queries.ListAuditCycles(c.Request().Context())
	if err != nil {
		h.logger.Errorw("Failed to list audit cycles", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list audit cycles")
	}

	response := make([]AuditCycleResponse, len(cycles))
	for i, cycle := range cycles {
		response[i] = convertToAuditCycleResponse(cycle)
	}

	return c.JSON(http.StatusOK, response)
}

// GetAuditCycle gets a specific audit cycle
// @Summary Get audit cycle
// @Tags audit-cycles
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Success 200 {object} AuditCycleResponse
// @Router /api/audit-cycles/{id} [get]
func (h *Handler) GetAuditCycle(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	cycle, err := h.store.Queries.GetAuditCycle(c.Request().Context(), cycleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Audit cycle not found")
		}
		h.logger.Errorw("Failed to get audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get audit cycle")
	}

	return c.JSON(http.StatusOK, convertToAuditCycleResponse(cycle))
}

// UpdateAuditCycle updates an audit cycle
// @Summary Update audit cycle
// @Tags audit-cycles
// @Accept json
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Param request body UpdateAuditCycleRequest true "Update details"
// @Success 200 {object} AuditCycleResponse
// @Router /api/audit-cycles/{id} [put]
func (h *Handler) UpdateAuditCycle(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	var req UpdateAuditCycleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Build update params
	// Get existing cycle to use as defaults
	existingCycle, err := h.store.Queries.GetAuditCycle(c.Request().Context(), cycleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Audit cycle not found")
		}
		h.logger.Errorw("Failed to get audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get audit cycle")
	}

	params := db.UpdateAuditCycleParams{
		ID:          cycleID,
		Name:        existingCycle.Name,
		Description: existingCycle.Description,
		StartDate:   existingCycle.StartDate,
		EndDate:     existingCycle.EndDate,
		Status:      existingCycle.Status,
	}

	if req.Name != nil {
		params.Name = *req.Name
	}

	if req.Description != nil {
		params.Description = req.Description
	}

	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		}
		params.StartDate = pgtype.Date{Time: startDate, Valid: true}
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		}
		params.EndDate = pgtype.Date{Time: endDate, Valid: true}
	}

	if req.Status != nil {
		params.Status = req.Status
	}

	cycle, err := h.store.Queries.UpdateAuditCycle(c.Request().Context(), params)
	if err != nil {
		h.logger.Errorw("Failed to update audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update audit cycle")
	}

	return c.JSON(http.StatusOK, convertToAuditCycleResponse(cycle))
}

// DeleteAuditCycle deletes an audit cycle
// @Summary Delete audit cycle
// @Tags audit-cycles
// @Param id path string true "Audit Cycle ID"
// @Success 204
// @Router /api/audit-cycles/{id} [delete]
func (h *Handler) DeleteAuditCycle(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	err = h.store.Queries.DeleteAuditCycle(c.Request().Context(), cycleID)
	if err != nil {
		h.logger.Errorw("Failed to delete audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete audit cycle")
	}

	return c.NoContent(http.StatusNoContent)
}

// AddClientToAuditCycle adds a client to an audit cycle
// @Summary Add client to audit cycle
// @Tags audit-cycles
// @Accept json
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Param request body AddClientToAuditCycleRequest true "Client details"
// @Success 201 {object} AuditCycleClientResponse
// @Router /api/audit-cycles/{id}/clients [post]
func (h *Handler) AddClientToAuditCycle(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	var req AddClientToAuditCycleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid client ID")
	}

	cycleClient, err := h.store.Queries.AddClientToAuditCycle(c.Request().Context(), db.AddClientToAuditCycleParams{
		AuditCycleID: cycleID,
		ClientID:     clientID,
	})
	if err != nil {
		h.logger.Errorw("Failed to add client to audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add client to audit cycle")
	}

	// Get client details
	clients, err := h.store.Queries.GetAuditCycleClients(c.Request().Context(), cycleID)
	if err != nil {
		h.logger.Errorw("Failed to get audit cycle clients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get client details")
	}

	// Find the newly added client
	for _, client := range clients {
		if client.ID == cycleClient.ID {
			return c.JSON(http.StatusCreated, convertToAuditCycleClientResponse(client))
		}
	}

	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve client details")
}

// GetAuditCycleClients gets all clients in an audit cycle
// @Summary Get audit cycle clients
// @Tags audit-cycles
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Success 200 {array} AuditCycleClientResponse
// @Router /api/audit-cycles/{id}/clients [get]
func (h *Handler) GetAuditCycleClients(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	clients, err := h.store.Queries.GetAuditCycleClients(c.Request().Context(), cycleID)
	if err != nil {
		h.logger.Errorw("Failed to get audit cycle clients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get audit cycle clients")
	}

	response := make([]AuditCycleClientResponse, len(clients))
	for i, client := range clients {
		response[i] = convertToAuditCycleClientResponse(client)
	}

	return c.JSON(http.StatusOK, response)
}

// RemoveClientFromAuditCycle removes a client from an audit cycle
// @Summary Remove client from audit cycle
// @Tags audit-cycles
// @Param id path string true "Audit Cycle ID"
// @Param clientId path string true "Client ID"
// @Success 204
// @Router /api/audit-cycles/{id}/clients/{clientId} [delete]
func (h *Handler) RemoveClientFromAuditCycle(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid client ID")
	}

	err = h.store.Queries.RemoveClientFromAuditCycle(c.Request().Context(), db.RemoveClientFromAuditCycleParams{
		AuditCycleID: cycleID,
		ClientID:     clientID,
	})
	if err != nil {
		h.logger.Errorw("Failed to remove client from audit cycle", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove client from audit cycle")
	}

	return c.NoContent(http.StatusNoContent)
}

// AssignFrameworkToClient assigns a framework to a client in an audit cycle
// @Summary Assign framework to client
// @Tags audit-cycles
// @Accept json
// @Produce json
// @Param cycleClientId path string true "Audit Cycle Client ID"
// @Param request body AssignFrameworkRequest true "Framework details"
// @Success 201 {object} AuditCycleFrameworkResponse
// @Router /api/audit-cycles/clients/{cycleClientId}/frameworks [post]
func (h *Handler) AssignFrameworkToClient(c echo.Context) error {
	cycleClientID, err := uuid.Parse(c.Param("cycleClientId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid cycle client ID")
	}

	var req AssignFrameworkRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	frameworkID, err := uuid.Parse(req.FrameworkID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid framework ID")
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	status := "pending"
	params := db.AssignFrameworkToAuditCycleClientParams{
		AuditCycleClientID: cycleClientID,
		FrameworkID:        frameworkID,
		FrameworkName:      req.FrameworkName,
		AssignedBy:         pgtype.UUID{Bytes: userID, Valid: true},
		Status:             &status,
	}

	if req.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid due_date format. Use YYYY-MM-DD")
		}
		params.DueDate = pgtype.Date{Time: dueDate, Valid: true}
	}

	if req.AuditorID != nil {
		auditorID, err := uuid.Parse(*req.AuditorID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid auditor_id format")
		}
		params.AuditorID = pgtype.UUID{Bytes: auditorID, Valid: true}
	}

	framework, err := h.store.Queries.AssignFrameworkToAuditCycleClient(c.Request().Context(), params)
	if err != nil {
		h.logger.Errorw("Failed to assign framework", "error", err)
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "unique_framework_per_client_cycle") || 
		   strings.Contains(err.Error(), "duplicate key value") {
			return echo.NewHTTPError(http.StatusConflict, "This framework is already assigned to this client in this audit cycle")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign framework")
	}

	// Convert to response (simplified version without client details)
	response := AuditCycleFrameworkResponse{
		ID:                 framework.ID.String(),
		AuditCycleClientID: framework.AuditCycleClientID.String(),
		FrameworkID:        framework.FrameworkID.String(),
		FrameworkName:      framework.FrameworkName,
		Status:             *framework.Status,
		CreatedAt:          framework.CreatedAt.Time,
		UpdatedAt:          framework.UpdatedAt.Time,
	}

	if framework.AssignedBy.Valid {
		assignedByUUID := uuid.UUID(framework.AssignedBy.Bytes)
		assignedBy := assignedByUUID.String()
		response.AssignedBy = &assignedBy
	}

	if framework.AssignedAt.Valid {
		response.AssignedAt = framework.AssignedAt.Time
	}

	if framework.DueDate.Valid {
		dueDate := framework.DueDate.Time.Format("2006-01-02")
		response.DueDate = &dueDate
	}

	if framework.AuditorID.Valid {
		auditorUUID := uuid.UUID(framework.AuditorID.Bytes)
		auditorID := auditorUUID.String()
		response.AuditorID = &auditorID
	}

	return c.JSON(http.StatusCreated, response)
}

// GetAuditCycleFrameworks gets all frameworks in an audit cycle
// @Summary Get audit cycle frameworks
// @Tags audit-cycles
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Success 200 {array} AuditCycleFrameworkResponse
// @Router /api/audit-cycles/{id}/frameworks [get]
func (h *Handler) GetAuditCycleFrameworks(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	frameworks, err := h.store.Queries.GetAuditCycleFrameworks(c.Request().Context(), cycleID)
	if err != nil {
		h.logger.Errorw("Failed to get audit cycle frameworks", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get audit cycle frameworks")
	}

	response := make([]AuditCycleFrameworkResponse, len(frameworks))
	for i, fw := range frameworks {
		response[i] = AuditCycleFrameworkResponse{
			ID:                 fw.ID.String(),
			AuditCycleClientID: fw.AuditCycleClientID.String(),
			FrameworkID:        fw.FrameworkID.String(),
			FrameworkName:      fw.FrameworkName,
			ClientID:           fw.ClientID.String(),
			ClientName:         fw.ClientName,
			AssignedAt:         fw.AssignedAt.Time,
			Status:             *fw.Status,
			CreatedAt:          fw.CreatedAt.Time,
			UpdatedAt:          fw.UpdatedAt.Time,
		}

		if fw.AssignedBy.Valid {
			assignedByUUID := uuid.UUID(fw.AssignedBy.Bytes)
			assignedBy := assignedByUUID.String()
			response[i].AssignedBy = &assignedBy
		}

		if fw.DueDate.Valid {
			dueDate := fw.DueDate.Time.Format("2006-01-02")
			response[i].DueDate = &dueDate
		}

		if fw.AuditorID.Valid {
			auditorUUID := uuid.UUID(fw.AuditorID.Bytes)
			auditorID := auditorUUID.String()
			response[i].AuditorID = &auditorID
		}
	}

	return c.JSON(http.StatusOK, response)
}

// GetAuditCycleStats gets statistics for an audit cycle
// @Summary Get audit cycle statistics
// @Tags audit-cycles
// @Produce json
// @Param id path string true "Audit Cycle ID"
// @Success 200 {object} AuditCycleStatsResponse
// @Router /api/audit-cycles/{id}/stats [get]
func (h *Handler) GetAuditCycleStats(c echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid audit cycle ID")
	}

	stats, err := h.store.Queries.GetAuditCycleStats(c.Request().Context(), cycleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Audit cycle not found")
		}
		h.logger.Errorw("Failed to get audit cycle stats", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get audit cycle stats")
	}

	var status string
	if stats.Status != nil {
		status = *stats.Status
	}

	response := AuditCycleStatsResponse{
		ID:                   stats.ID.String(),
		Name:                 stats.Name,
		Status:               status,
		TotalClients:         stats.TotalClients,
		TotalFrameworks:      stats.TotalFrameworks,
		CompletedFrameworks:  stats.CompletedFrameworks,
		InProgressFrameworks: stats.InProgressFrameworks,
		PendingFrameworks:    stats.PendingFrameworks,
		OverdueFrameworks:    stats.OverdueFrameworks,
	}

	return c.JSON(http.StatusOK, response)
}

// ============================================================================
// Helper Functions
// ============================================================================

func convertToAuditCycleResponse(cycle db.AuditCycle) AuditCycleResponse {
	var status string
	if cycle.Status != nil {
		status = *cycle.Status
	}

	response := AuditCycleResponse{
		ID:        cycle.ID.String(),
		Name:      cycle.Name,
		Status:    status,
		CreatedAt: cycle.CreatedAt.Time,
		UpdatedAt: cycle.UpdatedAt.Time,
	}

	if cycle.Description != nil {
		response.Description = *cycle.Description
	}

	if cycle.StartDate.Valid {
		response.StartDate = cycle.StartDate.Time.Format("2006-01-02")
	}

	if cycle.EndDate.Valid {
		response.EndDate = cycle.EndDate.Time.Format("2006-01-02")
	}

	if cycle.CreatedBy.Valid {
		createdByUUID := uuid.UUID(cycle.CreatedBy.Bytes)
		createdBy := createdByUUID.String()
		response.CreatedBy = &createdBy
	}

	return response
}

func convertToAuditCycleClientResponse(client db.GetAuditCycleClientsRow) AuditCycleClientResponse {
	var clientStatus string
	if client.ClientStatus.Valid {
		clientStatus = string(client.ClientStatus.ClientStatusEnum)
	}

	return AuditCycleClientResponse{
		ID:           client.ID.String(),
		AuditCycleID: client.AuditCycleID.String(),
		ClientID:     client.ClientID.String(),
		ClientName:   client.ClientName,
		POCEmail:     client.PocEmail,
		ClientStatus: clientStatus,
		CreatedAt:    client.CreatedAt.Time,
	}
}

func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userIDVal := c.Get("user_id")
	if userIDVal == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// The auth middleware sets user_id as uuid.UUID type
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID type")
	}

	return userID, nil
}
