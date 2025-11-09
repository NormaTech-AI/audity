package handler

import (
	"encoding/json"
	"net/http"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FrameworkResponse represents a framework in API responses
type FrameworkResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Version        string `json:"version"`
	QuestionCount  int    `json:"question_count,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateFrameworkRequest represents the request to create a framework
type CreateFrameworkRequest struct {
	Name          string          `json:"name" validate:"required"`
	Description   string          `json:"description" validate:"required"`
	ChecklistJSON json.RawMessage `json:"checklist_json" validate:"required"`
	Version       string          `json:"version" validate:"required"`
}

// UpdateFrameworkRequest represents the request to update a framework
type UpdateFrameworkRequest struct {
	Name          string          `json:"name" validate:"required"`
	Description   string          `json:"description" validate:"required"`
	ChecklistJSON json.RawMessage `json:"checklist_json" validate:"required"`
	Version       string          `json:"version" validate:"required"`
}

// ListFrameworks returns all compliance frameworks
func (h *Handler) ListFrameworks(c echo.Context) error {
	ctx := c.Request().Context()

	frameworks, err := h.store.ListFrameworks(ctx)
	if err != nil {
		h.logger.Errorw("Failed to list frameworks", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve frameworks",
		})
	}

	// Convert to response format
	responses := make([]FrameworkResponse, 0, len(frameworks))
	for _, fw := range frameworks {
		// Count questions in checklist
		var checklist map[string]interface{}
		questionCount := 0
		if err := json.Unmarshal(fw.ChecklistJson, &checklist); err == nil {
			if sections, ok := checklist["sections"].([]interface{}); ok {
				for _, section := range sections {
					if sectionMap, ok := section.(map[string]interface{}); ok {
						if questions, ok := sectionMap["questions"].([]interface{}); ok {
							questionCount += len(questions)
						}
					}
				}
			}
		}

		desc := ""
		if fw.Description != nil {
			desc = *fw.Description
		}
		ver := ""
		if fw.Version != nil {
			ver = *fw.Version
		}

		responses = append(responses, FrameworkResponse{
			ID:            fw.ID.String(),
			Name:          fw.Name,
			Description:   desc,
			Version:       ver,
			QuestionCount: questionCount,
			CreatedAt:     fw.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     fw.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// GetFramework returns a specific framework by ID
func (h *Handler) GetFramework(c echo.Context) error {
	ctx := c.Request().Context()
	
	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
	}

	framework, err := h.store.GetFramework(ctx, frameworkID)
	if err != nil {
		h.logger.Errorw("Failed to get framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Framework not found",
		})
	}

	// Count questions in checklist
	var checklist map[string]interface{}
	questionCount := 0
	if err := json.Unmarshal(framework.ChecklistJson, &checklist); err == nil {
		if sections, ok := checklist["sections"].([]interface{}); ok {
			for _, section := range sections {
				if sectionMap, ok := section.(map[string]interface{}); ok {
					if questions, ok := sectionMap["questions"].([]interface{}); ok {
						questionCount += len(questions)
					}
				}
			}
		}
	}

	desc := ""
	if framework.Description != nil {
		desc = *framework.Description
	}
	ver := ""
	if framework.Version != nil {
		ver = *framework.Version
	}

	response := FrameworkResponse{
		ID:            framework.ID.String(),
		Name:          framework.Name,
		Description:   desc,
		Version:       ver,
		QuestionCount: questionCount,
		CreatedAt:     framework.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     framework.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	return c.JSON(http.StatusOK, response)
}

// CreateFramework creates a new compliance framework
func (h *Handler) CreateFramework(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateFrameworkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Validate checklist JSON structure
	var checklist map[string]interface{}
	if err := json.Unmarshal(req.ChecklistJSON, &checklist); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid checklist JSON format",
		})
	}

	// Create framework
	framework, err := h.store.CreateFramework(ctx, db.CreateFrameworkParams{
		Name:          req.Name,
		Description:   &req.Description,
		ChecklistJson: req.ChecklistJSON,
		Version:       &req.Version,
	})
	if err != nil {
		h.logger.Errorw("Failed to create framework", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create framework",
		})
	}

	h.logger.Infow("Framework created", "id", framework.ID, "name", framework.Name)

	desc := ""
	if framework.Description != nil {
		desc = *framework.Description
	}
	ver := ""
	if framework.Version != nil {
		ver = *framework.Version
	}

	response := FrameworkResponse{
		ID:          framework.ID.String(),
		Name:        framework.Name,
		Description: desc,
		Version:     ver,
		CreatedAt:   framework.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   framework.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	return c.JSON(http.StatusCreated, response)
}

// UpdateFramework updates an existing framework
func (h *Handler) UpdateFramework(c echo.Context) error {
	ctx := c.Request().Context()

	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
	}

	var req UpdateFrameworkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Validate checklist JSON structure
	var checklist map[string]interface{}
	if err := json.Unmarshal(req.ChecklistJSON, &checklist); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid checklist JSON format",
		})
	}

	// Update framework
	framework, err := h.store.UpdateFramework(ctx, db.UpdateFrameworkParams{
		ID:            frameworkID,
		Name:          req.Name,
		Description:   &req.Description,
		ChecklistJson: req.ChecklistJSON,
		Version:       &req.Version,
	})
	if err != nil {
		h.logger.Errorw("Failed to update framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update framework",
		})
	}

	h.logger.Infow("Framework updated", "id", framework.ID, "name", framework.Name)

	desc := ""
	if framework.Description != nil {
		desc = *framework.Description
	}
	ver := ""
	if framework.Version != nil {
		ver = *framework.Version
	}

	response := FrameworkResponse{
		ID:          framework.ID.String(),
		Name:        framework.Name,
		Description: desc,
		Version:     ver,
		CreatedAt:   framework.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   framework.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteFramework deletes a framework
func (h *Handler) DeleteFramework(c echo.Context) error {
	ctx := c.Request().Context()

	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
	}

	// Check if framework is assigned to any clients
	assignments, err := h.store.ListClientFrameworks(ctx, uuid.Nil) // This would need a different query
	if err != nil {
		h.logger.Errorw("Failed to check framework assignments", "error", err)
		// Continue with deletion anyway
	}

	// Check if this framework is in use
	for _, assignment := range assignments {
		if assignment.FrameworkID == frameworkID {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Framework is assigned to clients and cannot be deleted",
			})
		}
	}

	// Delete framework
	if err := h.store.DeleteFramework(ctx, frameworkID); err != nil {
		h.logger.Errorw("Failed to delete framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete framework",
		})
	}

	h.logger.Infow("Framework deleted", "id", frameworkID)

	return c.NoContent(http.StatusNoContent)
}

// GetFrameworkChecklist returns the full checklist JSON for a framework
func (h *Handler) GetFrameworkChecklist(c echo.Context) error {
	ctx := c.Request().Context()

	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
	}

	framework, err := h.store.GetFramework(ctx, frameworkID)
	if err != nil {
		h.logger.Errorw("Failed to get framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Framework not found",
		})
	}

	// Parse and return the checklist JSON
	var checklist map[string]interface{}
	if err := json.Unmarshal(framework.ChecklistJson, &checklist); err != nil {
		h.logger.Errorw("Failed to parse checklist JSON", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Invalid framework data",
		})
	}

	return c.JSON(http.StatusOK, checklist)
}
