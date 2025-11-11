package handler

import (
	"net/http"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// ClientAuditListResponse represents an audit for client view
type ClientAuditListResponse struct {
	ID              string  `json:"id"`
	FrameworkID     string  `json:"framework_id"`
	FrameworkName   string  `json:"framework_name"`
	DueDate         string  `json:"due_date"`
	Status          string  `json:"status"`
	TotalQuestions  int64   `json:"total_questions"`
	AnsweredCount   int64   `json:"answered_count"`
	ProgressPercent float64 `json:"progress_percent"`
	CreatedAt       string  `json:"created_at"`
}

// ClientQuestionResponse represents a question with submission for client view
type ClientQuestionResponse struct {
	ID               string  `json:"id"`
	Section          string  `json:"section"`
	QuestionNumber   string  `json:"question_number"`
	QuestionText     string  `json:"question_text"`
	QuestionType     string  `json:"question_type"`
	HelpText         *string `json:"help_text"`
	IsMandatory      bool    `json:"is_mandatory"`
	DisplayOrder     int32   `json:"display_order"`
	SubmissionID     *string `json:"submission_id"`
	AnswerValue      *string `json:"answer_value"`
	AnswerText       *string `json:"answer_text"`
	Explanation      *string `json:"explanation"`
	SubmissionStatus *string `json:"submission_status"`
	SubmittedAt      *string `json:"submitted_at"`
	SubmittedBy      *string `json:"submitted_by"`
	IsAssignedToMe   bool    `json:"is_assigned_to_me"`
}

// ClientSubmissionRequest represents a submission payload from client
type ClientSubmissionRequest struct {
	QuestionID  string  `json:"question_id" validate:"required"`
	AnswerValue *string `json:"answer_value"`
	AnswerText  *string `json:"answer_text"`
	Explanation string  `json:"explanation" validate:"required"`
}

// ListClientAuditsView returns all audits for the authenticated client user
func (h *Handler) ListClientAuditsView(c echo.Context) error {
	ctx := c.Request().Context()

	// Get client ID from authenticated user
	clientID, err := getClientIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Client ID not found in context",
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

	// Convert to response format with progress
	responses := make([]ClientAuditListResponse, 0, len(audits))
	for _, audit := range audits {
		// Get audit progress
		progress, err := clientQueries.GetAuditProgress(ctx, audit.ID)
		if err != nil {
			h.logger.Warnw("Failed to get audit progress", "error", err, "audit_id", audit.ID)
			progress = clientdb.GetAuditProgressRow{
				TotalQuestions: 0,
				ApprovedCount:  0,
			}
		}

		progressPercent := 0.0
		if progress.TotalQuestions > 0 {
			progressPercent = float64(progress.ApprovedCount) / float64(progress.TotalQuestions) * 100
		}

		dueDate, _ := audit.DueDate.Value()
		createdAt, _ := audit.CreatedAt.Value()

		responses = append(responses, ClientAuditListResponse{
			ID:              audit.ID.String(),
			FrameworkID:     audit.FrameworkID.String(),
			FrameworkName:   audit.FrameworkName,
			DueDate:         dueDate.(string),
			Status:          string(audit.Status),
			TotalQuestions:  progress.TotalQuestions,
			AnsweredCount:   progress.ApprovedCount,
			ProgressPercent: progressPercent,
			CreatedAt:       createdAt.(string),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// GetClientAuditDetailView returns detailed audit information with questions for client
func (h *Handler) GetClientAuditDetailView(c echo.Context) error {
	ctx := c.Request().Context()

	// Get audit ID from path
	auditIDStr := c.Param("auditId")
	auditID, err := uuid.Parse(auditIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid audit ID",
		})
	}

	// Get client ID and user ID from authenticated user
	clientID, err := getClientIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Client ID not found in context",
		})
	}

	userID, err := getUserIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User ID not found in context",
		})
	}

	// Check if user is POC (can see all questions) or stakeholder (only assigned)
	isPOC, err := isUserPOCRole(c)
	if err != nil {
		h.logger.Warnw("Failed to determine user role", "error", err)
		isPOC = false
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Get audit details
	audit, err := clientQueries.GetAuditByID(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get audit", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Audit not found",
		})
	}

	// Get questions with role-based filtering
	questions, err := clientQueries.ListQuestionsForUser(ctx, clientdb.ListQuestionsForUserParams{
		AuditID:    auditID,
		Column2:    isPOC,
		AssignedTo: userID,
	})
	if err != nil {
		h.logger.Errorw("Failed to list questions", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve questions",
		})
	}

	// Convert questions to response format
	questionResponses := make([]ClientQuestionResponse, 0, len(questions))
	for _, q := range questions {
		var submissionID, answerValue, answerText, explanation, submissionStatus, submittedAt, submittedBy *string

		if q.SubmissionID.Valid {
			sid := q.SubmissionID.Bytes
			sidUUID, _ := uuid.FromBytes(sid[:])
			sidStr := sidUUID.String()
			submissionID = &sidStr
		}

		if q.AnswerValue.Valid {
			av := string(q.AnswerValue.AnswerValueEnum)
			answerValue = &av
		}

		if q.AnswerText != nil {
			answerText = q.AnswerText
		}

		if q.Explanation != nil {
			explanation = q.Explanation
		}

		if q.SubmissionStatus.Valid {
			ss := string(q.SubmissionStatus.SubmissionStatusEnum)
			submissionStatus = &ss
		}

		if q.SubmittedAt.Valid {
			sat, _ := q.SubmittedAt.Value()
			if sat != nil {
				satStr := sat.(string)
				submittedAt = &satStr
			}
		}

		if q.SubmittedBy.Valid {
			sb := q.SubmittedBy.Bytes
			sbUUID, _ := uuid.FromBytes(sb[:])
			sbStr := sbUUID.String()
			submittedBy = &sbStr
		}

		isAssignedToMe := false
		if q.AssignedUserID.Valid {
			assignedID := q.AssignedUserID.Bytes
			assignedUUID, _ := uuid.FromBytes(assignedID[:])
			isAssignedToMe = assignedUUID == userID
		}

		questionResponses = append(questionResponses, ClientQuestionResponse{
			ID:               q.ID.String(),
			Section:          q.Section,
			QuestionNumber:   q.QuestionNumber,
			QuestionText:     q.QuestionText,
			QuestionType:     string(q.QuestionType),
			HelpText:         q.HelpText,
			IsMandatory:      q.IsMandatory,
			DisplayOrder:     q.DisplayOrder,
			SubmissionID:     submissionID,
			AnswerValue:      answerValue,
			AnswerText:       answerText,
			Explanation:      explanation,
			SubmissionStatus: submissionStatus,
			SubmittedAt:      submittedAt,
			SubmittedBy:      submittedBy,
			IsAssignedToMe:   isAssignedToMe,
		})
	}

	dueDate, _ := audit.DueDate.Value()
	createdAt, _ := audit.CreatedAt.Value()

	response := map[string]interface{}{
		"id":             audit.ID.String(),
		"framework_id":   audit.FrameworkID.String(),
		"framework_name": audit.FrameworkName,
		"due_date":       dueDate,
		"status":         string(audit.Status),
		"created_at":     createdAt,
		"questions":      questionResponses,
	}

	return c.JSON(http.StatusOK, response)
}

// SaveClientSubmission creates or updates a submission (saves as draft)
func (h *Handler) SaveClientSubmission(c echo.Context) error {
	ctx := c.Request().Context()

	var req ClientSubmissionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Get client ID and user ID from authenticated user
	clientID, err := getClientIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Client ID not found in context",
		})
	}

	userID, err := getUserIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User ID not found in context",
		})
	}

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid question ID",
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

	// Check if submission already exists
	existingSubmission, err := clientQueries.GetSubmissionByQuestionID(ctx, questionID)
	
	var submission clientdb.Submission
	
	if err == nil && existingSubmission.ID != uuid.Nil {
		// Update existing submission
		var answerValue clientdb.NullAnswerValueEnum
		if req.AnswerValue != nil {
			answerValue = clientdb.NullAnswerValueEnum{
				AnswerValueEnum: clientdb.AnswerValueEnum(*req.AnswerValue),
				Valid:           true,
			}
		}

		var answerText pgtype.Text
		if req.AnswerText != nil {
			answerText = pgtype.Text{String: *req.AnswerText, Valid: true}
		}

		submission, err = clientQueries.UpdateSubmissionAnswer(ctx, clientdb.UpdateSubmissionAnswerParams{
			ID:          existingSubmission.ID,
			AnswerValue: answerValue,
			AnswerText:  &answerText.String,
			Explanation: req.Explanation,
		})
		if err != nil {
			h.logger.Errorw("Failed to update submission", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update submission",
			})
		}
	} else {
		// Create new submission
		var answerValue clientdb.NullAnswerValueEnum
		if req.AnswerValue != nil {
			answerValue = clientdb.NullAnswerValueEnum{
				AnswerValueEnum: clientdb.AnswerValueEnum(*req.AnswerValue),
				Valid:           true,
			}
		}

		var answerText pgtype.Text
		var answerTextPtr *string
		if req.AnswerText != nil {
			answerText = pgtype.Text{String: *req.AnswerText, Valid: true}
			answerTextPtr = &answerText.String
		}

		submission, err = clientQueries.CreateSubmission(ctx, clientdb.CreateSubmissionParams{
			QuestionID:  questionID,
			SubmittedBy: userID,
			AnswerValue: answerValue,
			AnswerText:  answerTextPtr,
			Explanation: req.Explanation,
			Status:      clientdb.SubmissionStatusEnumInProgress,
		})
		if err != nil {
			h.logger.Errorw("Failed to create submission", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create submission",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":          submission.ID.String(),
		"question_id": submission.QuestionID.String(),
		"status":      string(submission.Status),
		"message":     "Submission saved successfully",
	})
}

// SubmitClientAnswer submits a saved answer for review
func (h *Handler) SubmitClientAnswer(c echo.Context) error {
	ctx := c.Request().Context()

	submissionIDStr := c.Param("submissionId")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
		})
	}

	// Get client ID from authenticated user
	clientID, err := getClientIDFromUser(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Client ID not found in context",
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

	// Submit the submission
	submission, err := clientQueries.SubmitSubmission(ctx, submissionID)
	if err != nil {
		h.logger.Errorw("Failed to submit submission", "error", err, "submission_id", submissionID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to submit answer",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":      submission.ID.String(),
		"status":  string(submission.Status),
		"message": "Answer submitted successfully for review",
	})
}

// Helper functions for client audit view

func getClientIDFromUser(c echo.Context) (uuid.UUID, error) {
	user := c.Get("user")
	if user == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}
	
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}
	
	clientIDStr, ok := userMap["client_id"].(string)
	if !ok || clientIDStr == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Client ID not found")
	}
	
	return uuid.Parse(clientIDStr)
}

func getUserIDFromUser(c echo.Context) (uuid.UUID, error) {
	user := c.Get("user")
	if user == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}
	
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}
	
	userIDStr, ok := userMap["user_id"].(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}
	
	return uuid.Parse(userIDStr)
}

func isUserPOCRole(c echo.Context) (bool, error) {
	user := c.Get("user")
	if user == nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "User not found in context")
	}
	
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return false, echo.NewHTTPError(http.StatusUnauthorized, "Invalid user data")
	}
	
	designation, ok := userMap["designation"].(string)
	if !ok {
		return false, nil
	}
	
	// POC users can see all questions
	return designation == "poc_client" || designation == "poc_internal", nil
}
