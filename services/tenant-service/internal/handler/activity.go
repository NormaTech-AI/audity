package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ActivityLogResponse represents an activity log entry in API responses
type ActivityLogResponse struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	UserEmail  string                 `json:"user_email"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
	CreatedAt  string                 `json:"created_at"`
}

// CreateActivityLog creates a new activity log entry
func (h *Handler) CreateActivityLog(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	var req struct {
		Action     string                 `json:"action" validate:"required"`
		EntityType string                 `json:"entity_type" validate:"required"`
		EntityID   string                 `json:"entity_id" validate:"required"`
		Details    map[string]interface{} `json:"details"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid entity ID",
		})
	}

	// Get user info from context
	userID := c.Get("user_id").(string)
	userEmail := c.Get("user_email").(string)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
		})
	}

	// Get client IP and user agent
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Marshal details to JSON
	var detailsJSON []byte
	if req.Details != nil {
		detailsJSON, err = json.Marshal(req.Details)
		if err != nil {
			h.logger.Errorw("Failed to marshal details", "error", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid details format",
			})
		}
	} else {
		detailsJSON = []byte("{}")
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Create activity log
	activityLog, err := clientQueries.CreateActivityLog(ctx, clientdb.CreateActivityLogParams{
		UserID:     userUUID,
		UserEmail:  userEmail,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   entityID,
		Details:    detailsJSON,
		IpAddress:  &ipAddress,
		UserAgent:  &userAgent,
	})
	if err != nil {
		h.logger.Errorw("Failed to create activity log", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create activity log",
		})
	}

	h.logger.Infow("Activity log created", 
		"log_id", activityLog.ID, 
		"action", req.Action,
		"entity_type", req.EntityType,
		"client_id", clientID)

	response := buildActivityLogResponse(activityLog)

	return c.JSON(http.StatusCreated, response)
}

// ListActivityLogs lists activity logs with pagination
func (h *Handler) ListActivityLogs(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	// Get pagination parameters
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// List activity logs
	logs, err := clientQueries.ListActivityLogs(ctx, clientdb.ListActivityLogsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Errorw("Failed to list activity logs", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve activity logs",
		})
	}

	// Convert to response format
	responses := make([]ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, buildActivityLogResponse(log))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
		"count":  len(responses),
	})
}

// ListActivityLogsByUser lists activity logs for a specific user
func (h *Handler) ListActivityLogsByUser(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	// Get pagination parameters
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// List activity logs by user
	logs, err := clientQueries.ListActivityLogsByUser(ctx, clientdb.ListActivityLogsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Errorw("Failed to list activity logs by user", "error", err, "user_id", userID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve activity logs",
		})
	}

	// Convert to response format
	responses := make([]ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, buildActivityLogResponse(log))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":   responses,
		"limit":  limit,
		"offset": offset,
		"count":  len(responses),
	})
}

// ListActivityLogsByEntity lists activity logs for a specific entity
func (h *Handler) ListActivityLogsByEntity(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	entityType := c.QueryParam("entity_type")
	if entityType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Entity type is required",
		})
	}

	entityID, err := uuid.Parse(c.QueryParam("entity_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid entity ID",
		})
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// List activity logs by entity
	logs, err := clientQueries.ListActivityLogsByEntity(ctx, clientdb.ListActivityLogsByEntityParams{
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		h.logger.Errorw("Failed to list activity logs by entity", "error", err, "entity_type", entityType, "entity_id", entityID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve activity logs",
		})
	}

	// Convert to response format
	responses := make([]ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, buildActivityLogResponse(log))
	}

	return c.JSON(http.StatusOK, responses)
}

// GetRecentActivity gets recent activity logs
func (h *Handler) GetRecentActivity(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	// Get limit parameter (default 20, max 50)
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Get recent activity
	logs, err := clientQueries.GetRecentActivity(ctx, int32(limit))
	if err != nil {
		h.logger.Errorw("Failed to get recent activity", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve recent activity",
		})
	}

	// Convert to response format
	responses := make([]ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, buildActivityLogResponse(log))
	}

	return c.JSON(http.StatusOK, responses)
}

// Helper function to build activity log response
func buildActivityLogResponse(log clientdb.ActivityLog) ActivityLogResponse {
	var details map[string]interface{}
	if len(log.Details) > 0 {
		json.Unmarshal(log.Details, &details)
	}

	return ActivityLogResponse{
		ID:         log.ID.String(),
		UserID:     log.UserID.String(),
		UserEmail:  log.UserEmail,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID.String(),
		Details:    details,
		IPAddress:  log.IpAddress,
		UserAgent:  log.UserAgent,
		CreatedAt:  log.CreatedAt.Time.Format(time.RFC3339),
	}
}
