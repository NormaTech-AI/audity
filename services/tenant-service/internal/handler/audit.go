package handler

import (
	"net/http"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// AuditResponse represents an audit in API responses
type AuditResponse struct {
	ID              string  `json:"id"`
	FrameworkID     string  `json:"framework_id"`
	FrameworkName   string  `json:"framework_name"`
	AssignedBy      string  `json:"assigned_by"`
	AssignedTo      *string `json:"assigned_to"`
	DueDate         string  `json:"due_date"`
	Status          string  `json:"status"`
	TotalQuestions  int     `json:"total_questions"`
	AnsweredCount   int     `json:"answered_count"`
	ApprovedCount   int     `json:"approved_count"`
	ProgressPercent float64 `json:"progress_percent"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// AuditDetailResponse includes questions
type AuditDetailResponse struct {
	AuditResponse
	Questions []QuestionWithSubmissionResponse `json:"questions"`
}

// QuestionWithSubmissionResponse represents a question with its submission status
type QuestionWithSubmissionResponse struct {
	ID             string  `json:"id"`
	Section        string  `json:"section"`
	QuestionNumber string  `json:"question_number"`
	QuestionText   string  `json:"question_text"`
	QuestionType   string  `json:"question_type"`
	HelpText       *string `json:"help_text"`
	IsMandatory    bool    `json:"is_mandatory"`
	DisplayOrder   int32   `json:"display_order"`
	SubmissionID   *string `json:"submission_id"`
	Answer         *string `json:"answer"`
	Status         *string `json:"status"`
	SubmittedAt    *string `json:"submitted_at"`
}

// UpdateAuditRequest represents the request to update an audit
type UpdateAuditRequest struct {
	AssignedTo *string `json:"assigned_to"`
	DueDate    *string `json:"due_date"`
	Status     *string `json:"status"`
}

// ListClientAudits returns all audits for a specific client
func (h *Handler) ListClientAudits(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
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

	// List all audits
	audits, err := clientQueries.ListAudits(ctx)
	if err != nil {
		h.logger.Errorw("Failed to list audits", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve audits",
		})
	}

	// Convert to response format
	responses := make([]AuditResponse, 0, len(audits))
	for _, audit := range audits {
		// Get audit progress
		progress, err := clientQueries.GetAuditProgress(ctx, audit.ID)
		if err != nil {
			h.logger.Warnw("Failed to get audit progress", "error", err, "audit_id", audit.ID)
			progress = clientdb.GetAuditProgressRow{
				TotalQuestions: 0,
				SubmittedCount: 0,
				ApprovedCount:  0,
			}
		}

		var assignedTo *string
		if audit.AssignedTo.Valid {
			assignedToStr := uuid.UUID(audit.AssignedTo.Bytes).String()
			assignedTo = &assignedToStr
		}

		progressPercent := 0.0
		if progress.TotalQuestions > 0 {
			progressPercent = (float64(progress.SubmittedCount) / float64(progress.TotalQuestions)) * 100
		}

		responses = append(responses, AuditResponse{
			ID:              audit.ID.String(),
			FrameworkID:     audit.FrameworkID.String(),
			FrameworkName:   audit.FrameworkName,
			AssignedBy:      audit.AssignedBy.String(),
			AssignedTo:      assignedTo,
			DueDate:         audit.DueDate.Time.Format("2006-01-02"),
			Status:          string(audit.Status),
			TotalQuestions:  int(progress.TotalQuestions),
			AnsweredCount:   int(progress.SubmittedCount),
			ApprovedCount:   int(progress.ApprovedCount),
			ProgressPercent: progressPercent,
			CreatedAt:       audit.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       audit.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// GetAudit returns a specific audit with all questions
func (h *Handler) GetAudit(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	auditID, err := uuid.Parse(c.Param("auditId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid audit ID",
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

	// Get audit
	audit, err := clientQueries.GetAuditByID(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get audit", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Audit not found",
		})
	}

	// Get audit progress
	progress, err := clientQueries.GetAuditProgress(ctx, auditID)
	if err != nil {
		h.logger.Warnw("Failed to get audit progress", "error", err, "audit_id", auditID)
		progress = clientdb.GetAuditProgressRow{
			TotalQuestions: 0,
			SubmittedCount: 0,
			ApprovedCount:  0,
		}
	}

	// Get questions with submissions
	questionsWithSubs, err := clientQueries.ListQuestionsWithSubmissions(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get questions", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve questions",
		})
	}

	// Convert questions to response format
	questions := make([]QuestionWithSubmissionResponse, 0, len(questionsWithSubs))
	for _, q := range questionsWithSubs {
		var helpText *string
		if q.HelpText != nil {
			helpText = q.HelpText
		}

		var submissionID *string
		if q.SubmissionID.Valid {
			sid := uuid.UUID(q.SubmissionID.Bytes).String()
			submissionID = &sid
		}

		// Answer data would come from a separate submission query
		// For now, we don't have the answer in this query
		var answer *string = nil

		var status *string
		if q.SubmissionStatus.Valid {
			st := string(q.SubmissionStatus.SubmissionStatusEnum)
			status = &st
		}

		var submittedAt *string
		if q.SubmittedAt.Valid {
			sa := q.SubmittedAt.Time.Format("2006-01-02T15:04:05Z")
			submittedAt = &sa
		}

		questions = append(questions, QuestionWithSubmissionResponse{
			ID:             q.ID.String(),
			Section:        q.Section,
			QuestionNumber: q.QuestionNumber,
			QuestionText:   q.QuestionText,
			QuestionType:   string(q.QuestionType),
			HelpText:       helpText,
			IsMandatory:    q.IsMandatory,
			DisplayOrder:   q.DisplayOrder,
			SubmissionID:   submissionID,
			Answer:         answer,
			Status:         status,
			SubmittedAt:    submittedAt,
		})
	}

	var assignedTo *string
	if audit.AssignedTo.Valid {
		assignedToStr := uuid.UUID(audit.AssignedTo.Bytes).String()
		assignedTo = &assignedToStr
	}

	progressPercent := 0.0
	if progress.TotalQuestions > 0 {
		progressPercent = (float64(progress.SubmittedCount) / float64(progress.TotalQuestions)) * 100
	}

	response := AuditDetailResponse{
		AuditResponse: AuditResponse{
			ID:              audit.ID.String(),
			FrameworkID:     audit.FrameworkID.String(),
			FrameworkName:   audit.FrameworkName,
			AssignedBy:      audit.AssignedBy.String(),
			AssignedTo:      assignedTo,
			DueDate:         audit.DueDate.Time.Format("2006-01-02"),
			Status:          string(audit.Status),
			TotalQuestions:  int(progress.TotalQuestions),
			AnsweredCount:   int(progress.SubmittedCount),
			ApprovedCount:   int(progress.ApprovedCount),
			ProgressPercent: progressPercent,
			CreatedAt:       audit.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       audit.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		},
		Questions: questions,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateAudit updates an audit (assignment, due date, status)
func (h *Handler) UpdateAudit(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	auditID, err := uuid.Parse(c.Param("auditId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid audit ID",
		})
	}

	var req UpdateAuditRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
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

	// Update assigned_to if provided
	if req.AssignedTo != nil {
		assignedToUUID, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid assigned_to UUID",
			})
		}

		_, err = clientQueries.UpdateAuditAssignee(ctx, clientdb.UpdateAuditAssigneeParams{
			ID:         auditID,
			AssignedTo: pgtype.UUID{Bytes: assignedToUUID, Valid: true},
		})
		if err != nil {
			h.logger.Errorw("Failed to update audit assignee", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update audit",
			})
		}
	}

	// Update status if provided
	if req.Status != nil {
		status := clientdb.AuditStatusEnum(*req.Status)
		_, err = clientQueries.UpdateAuditStatus(ctx, clientdb.UpdateAuditStatusParams{
			ID:     auditID,
			Status: status,
		})
		if err != nil {
			h.logger.Errorw("Failed to update audit status", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update audit",
			})
		}
	}

	// Get updated audit
	audit, err := clientQueries.GetAuditByID(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get updated audit", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve updated audit",
		})
	}

	// Get progress
	progress, err := clientQueries.GetAuditProgress(ctx, auditID)
	if err != nil {
		progress = clientdb.GetAuditProgressRow{
			TotalQuestions: 0,
			SubmittedCount: 0,
			ApprovedCount:  0,
		}
	}

	var assignedTo *string
	if audit.AssignedTo.Valid {
		assignedToStr := uuid.UUID(audit.AssignedTo.Bytes).String()
		assignedTo = &assignedToStr
	}

	progressPercent := 0.0
	if progress.TotalQuestions > 0 {
		progressPercent = (float64(progress.SubmittedCount) / float64(progress.TotalQuestions)) * 100
	}

	response := AuditResponse{
		ID:              audit.ID.String(),
		FrameworkID:     audit.FrameworkID.String(),
		FrameworkName:   audit.FrameworkName,
		AssignedBy:      audit.AssignedBy.String(),
		AssignedTo:      assignedTo,
		DueDate:         audit.DueDate.Time.Format("2006-01-02"),
		Status:          string(audit.Status),
		TotalQuestions:  int(progress.TotalQuestions),
		AnsweredCount:   int(progress.SubmittedCount),
		ApprovedCount:   int(progress.ApprovedCount),
		ProgressPercent: progressPercent,
		CreatedAt:       audit.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       audit.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	h.logger.Infow("Audit updated", "audit_id", auditID, "client_id", clientID)

	return c.JSON(http.StatusOK, response)
}
