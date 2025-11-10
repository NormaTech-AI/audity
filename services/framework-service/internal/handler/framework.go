package handler

import (
	"net/http"

	"github.com/NormaTech-AI/audity/services/framework-service/internal/db"
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

// FrameworkQuestionRequest represents a question in the request
type FrameworkQuestionRequest struct {
	SectionTitle       *string  `json:"section_title"`
	ControlID          string   `json:"control_id" validate:"required"`
	QuestionText       string   `json:"question_text" validate:"required"`
	HelpText           *string  `json:"help_text"`
	AcceptableEvidence []string `json:"acceptable_evidence"`
}

// CreateFrameworkRequest represents the request to create a framework
type CreateFrameworkRequest struct {
	Name        string                      `json:"name" validate:"required"`
	Description string                      `json:"description" validate:"required"`
	Version     string                      `json:"version" validate:"required"`
	Questions   []FrameworkQuestionRequest  `json:"questions" validate:"required,min=1"`
}

// UpdateFrameworkRequest represents the request to update a framework
type UpdateFrameworkRequest struct {
	Name        string                      `json:"name" validate:"required"`
	Description string                      `json:"description" validate:"required"`
	Version     string                      `json:"version" validate:"required"`
	Questions   []FrameworkQuestionRequest  `json:"questions" validate:"required,min=1"`
}

// ListFrameworks returns all compliance frameworks
// @Summary List all frameworks
// @Description Get a list of all compliance frameworks
// @Tags frameworks
// @Accept json
// @Produce json
// @Success 200 {array} FrameworkResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/frameworks [get]
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
		// Count questions from the new table
		count, err := h.store.CountFrameworkQuestions(ctx, fw.ID)
		if err != nil {
			h.logger.Errorw("Failed to count questions", "error", err, "framework_id", fw.ID)
			count = 0
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
			QuestionCount: int(count),
			CreatedAt:     fw.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     fw.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// GetFramework returns a specific framework by ID
// @Summary Get framework by ID
// @Description Get a specific compliance framework by its ID
// @Tags frameworks
// @Accept json
// @Produce json
// @Param id path string true "Framework ID"
// @Success 200 {object} FrameworkResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/frameworks/{id} [get]
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

	// Count questions from the new table
	count, err := h.store.CountFrameworkQuestions(ctx, frameworkID)
	if err != nil {
		h.logger.Errorw("Failed to count questions", "error", err, "framework_id", frameworkID)
		count = 0
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
		QuestionCount: int(count),
		CreatedAt:     framework.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     framework.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	return c.JSON(http.StatusOK, response)
}

// CreateFramework creates a new compliance framework
// @Summary Create a new framework
// @Description Create a new compliance framework
// @Tags frameworks
// @Accept json
// @Produce json
// @Param framework body CreateFrameworkRequest true "Framework data"
// @Success 201 {object} FrameworkResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/frameworks [post]
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

	framework, err := h.store.CreateFramework(ctx, db.CreateFrameworkParams{
		Name:          req.Name,
		Description:   &req.Description,
		Version:       &req.Version,
	})
	if err != nil {
		h.logger.Errorw("Failed to create framework", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create framework",
		})
	}

	// Create questions
	for _, q := range req.Questions {
		_, err := h.store.CreateFrameworkQuestion(ctx, db.CreateFrameworkQuestionParams{
			FrameworkID:        framework.ID,
			SectionTitle:       q.SectionTitle,
			ControlID:          q.ControlID,
			QuestionText:       q.QuestionText,
			HelpText:           q.HelpText,
			AcceptableEvidence: q.AcceptableEvidence,
		})
		if err != nil {
			h.logger.Errorw("Failed to create question", "error", err)
			// Rollback: delete the framework
			_ = h.store.DeleteFramework(ctx, framework.ID)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create framework questions",
			})
		}
	}

	h.logger.Infow("Framework created", "id", framework.ID, "name", framework.Name, "questions", len(req.Questions))

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
// @Summary Update a framework
// @Description Update an existing compliance framework
// @Tags frameworks
// @Accept json
// @Produce json
// @Param id path string true "Framework ID"
// @Param framework body UpdateFrameworkRequest true "Framework data"
// @Success 200 {object} FrameworkResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/frameworks/{id} [put]
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

	framework, err := h.store.UpdateFramework(ctx, db.UpdateFrameworkParams{
		ID:            frameworkID,
		Name:          req.Name,
		Description:   &req.Description,
		Version:       &req.Version,
	})
	if err != nil {
		h.logger.Errorw("Failed to update framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update framework",
		})
	}

	// Delete existing questions and create new ones
	if err := h.store.DeleteFrameworkQuestionsByFrameworkId(ctx, frameworkID); err != nil {
		h.logger.Errorw("Failed to delete existing questions", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update framework questions",
		})
	}

	// Create new questions
	for _, q := range req.Questions {
		_, err := h.store.CreateFrameworkQuestion(ctx, db.CreateFrameworkQuestionParams{
			FrameworkID:        framework.ID,
			SectionTitle:       q.SectionTitle,
			ControlID:          q.ControlID,
			QuestionText:       q.QuestionText,
			HelpText:           q.HelpText,
			AcceptableEvidence: q.AcceptableEvidence,
		})
		if err != nil {
			h.logger.Errorw("Failed to create question", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update framework questions",
			})
		}
	}

	h.logger.Infow("Framework updated", "id", framework.ID, "name", framework.Name, "questions", len(req.Questions))

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
// @Summary Delete a framework
// @Description Delete a compliance framework
// @Tags frameworks
// @Accept json
// @Produce json
// @Param id path string true "Framework ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/frameworks/{id} [delete]
func (h *Handler) DeleteFramework(c echo.Context) error {
	ctx := c.Request().Context()

	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
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
// @Summary Get framework checklist
// @Description Get the full checklist JSON for a framework
// @Tags frameworks
// @Accept json
// @Produce json
// @Param id path string true "Framework ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/frameworks/{id}/checklist [get]
func (h *Handler) GetFrameworkChecklist(c echo.Context) error {
	ctx := c.Request().Context()

	frameworkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid framework ID",
		})
	}

	// Get framework to verify it exists
	_, err = h.store.GetFramework(ctx, frameworkID)
	if err != nil {
		h.logger.Errorw("Failed to get framework", "error", err, "id", frameworkID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Framework not found",
		})
	}

	// Get questions from the new table
	questions, err := h.store.ListFrameworkQuestions(ctx, frameworkID)
	if err != nil {
		h.logger.Errorw("Failed to get framework questions", "error", err, "id", frameworkID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve framework questions",
		})
	}

	// Convert to response format
	type QuestionResponse struct {
		QuestionID         string   `json:"question_id"`
		SectionTitle       *string  `json:"section_title"`
		ControlID          string   `json:"control_id"`
		QuestionText       string   `json:"question_text"`
		HelpText           *string  `json:"help_text"`
		AcceptableEvidence []string `json:"acceptable_evidence"`
	}

	response := make([]QuestionResponse, 0, len(questions))
	for _, q := range questions {
		response = append(response, QuestionResponse{
			QuestionID:         q.QuestionID.String(),
			SectionTitle:       q.SectionTitle,
			ControlID:          q.ControlID,
			QuestionText:       q.QuestionText,
			HelpText:           q.HelpText,
			AcceptableEvidence: q.AcceptableEvidence,
		})
	}

	return c.JSON(http.StatusOK, response)
}
