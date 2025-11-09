package handler

import (
	"net/http"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// SubmissionResponse represents a submission in API responses
type SubmissionResponse struct {
	ID             string  `json:"id"`
	QuestionID     string  `json:"question_id"`
	SubmittedBy    string  `json:"submitted_by"`
	Answer         *string `json:"answer"`
	AnswerValue    *string `json:"answer_value"`
	Status         string  `json:"status"`
	Version        int32   `json:"version"`
	ReviewedBy     *string `json:"reviewed_by"`
	ReviewedAt     *string `json:"reviewed_at"`
	RejectionNotes *string `json:"rejection_notes"`
	SubmittedAt    *string `json:"submitted_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// CreateSubmissionRequest represents the request to create/update a submission
type CreateSubmissionRequest struct {
	QuestionID  string  `json:"question_id" validate:"required"`
	Answer      *string `json:"answer"`
	AnswerValue *string `json:"answer_value"`
}

// UpdateSubmissionRequest represents the request to update a submission answer
type UpdateSubmissionRequest struct {
	Answer      *string `json:"answer"`
	AnswerValue *string `json:"answer_value"`
}

// ReviewSubmissionRequest represents the request to review a submission
type ReviewSubmissionRequest struct {
	Action         string  `json:"action" validate:"required,oneof=approve reject refer"`
	RejectionNotes *string `json:"rejection_notes"`
}

// CreateOrUpdateSubmission creates or updates a draft submission
func (h *Handler) CreateOrUpdateSubmission(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	var req CreateSubmissionRequest
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

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid question ID",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Get("user_id").(string)
	submittedBy, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
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
	existingSubmissions, err := clientQueries.ListSubmissionsByUser(ctx, submittedBy)
	if err != nil {
		h.logger.Errorw("Failed to check existing submissions", "error", err)
	}

	var submission clientdb.Submission
	var isUpdate bool

	// Find existing in-progress submission for this question
	for _, sub := range existingSubmissions {
		if sub.QuestionID == questionID && sub.Status == clientdb.SubmissionStatusEnumInProgress {
			isUpdate = true
			
			// Convert answer_value to nullable enum
			var answerValue clientdb.NullAnswerValueEnum
			if req.AnswerValue != nil {
				answerValue = clientdb.NullAnswerValueEnum{
					AnswerValueEnum: clientdb.AnswerValueEnum(*req.AnswerValue),
					Valid:           true,
				}
			}

			// Update existing submission
			submission, err = clientQueries.UpdateSubmissionAnswer(ctx, clientdb.UpdateSubmissionAnswerParams{
				ID:          sub.ID,
				AnswerText:  req.Answer,
				AnswerValue: answerValue,
			})
			if err != nil {
				h.logger.Errorw("Failed to update submission", "error", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to update submission",
				})
			}
			break
		}
	}

	// Create new submission if not updating
	if !isUpdate {
		// Convert answer_value to nullable enum
		var answerValue clientdb.NullAnswerValueEnum
		if req.AnswerValue != nil {
			answerValue = clientdb.NullAnswerValueEnum{
				AnswerValueEnum: clientdb.AnswerValueEnum(*req.AnswerValue),
				Valid:           true,
			}
		}

		submission, err = clientQueries.CreateSubmission(ctx, clientdb.CreateSubmissionParams{
			QuestionID:  questionID,
			SubmittedBy: submittedBy,
			AnswerText:  req.Answer,
			AnswerValue: answerValue,
			Explanation: "", // Optional explanation
			Status:      clientdb.SubmissionStatusEnumInProgress,
		})
		if err != nil {
			h.logger.Errorw("Failed to create submission", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create submission",
			})
		}
	}

	response := buildSubmissionResponse(submission)

	statusCode := http.StatusOK
	if !isUpdate {
		statusCode = http.StatusCreated
	}

	h.logger.Infow("Submission saved", 
		"submission_id", submission.ID, 
		"question_id", questionID, 
		"client_id", clientID,
		"is_update", isUpdate)

	return c.JSON(statusCode, response)
}

// SubmitForReview submits a submission for review
func (h *Handler) SubmitForReview(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	submissionID, err := uuid.Parse(c.Param("submissionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
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
			"error": "Failed to submit submission",
		})
	}

	response := buildSubmissionResponse(submission)

	h.logger.Infow("Submission submitted for review", 
		"submission_id", submissionID, 
		"client_id", clientID)

	return c.JSON(http.StatusOK, response)
}

// ReviewSubmission allows auditors to review submissions
func (h *Handler) ReviewSubmission(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	submissionID, err := uuid.Parse(c.Param("submissionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
		})
	}

	var req ReviewSubmissionRequest
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

	// Get reviewer ID from context
	reviewerID := c.Get("user_id").(string)
	reviewedBy, err := uuid.Parse(reviewerID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
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

	var submission clientdb.Submission

	switch req.Action {
	case "approve":
		var reviewNotes *string
		if req.RejectionNotes != nil {
			reviewNotes = req.RejectionNotes
		}

		submission, err = clientQueries.ApproveSubmission(ctx, clientdb.ApproveSubmissionParams{
			ID:          submissionID,
			ReviewedBy:  pgtype.UUID{Bytes: reviewedBy, Valid: true},
			ReviewNotes: reviewNotes,
		})

	case "reject":
		if req.RejectionNotes == nil || *req.RejectionNotes == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Rejection notes are required when rejecting a submission",
			})
		}

		submission, err = clientQueries.RejectSubmission(ctx, clientdb.RejectSubmissionParams{
			ID:              submissionID,
			ReviewedBy:      pgtype.UUID{Bytes: reviewedBy, Valid: true},
			RejectionReason: req.RejectionNotes,
		})

	case "refer":
		if req.RejectionNotes == nil || *req.RejectionNotes == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Notes are required when referring a submission",
			})
		}

		submission, err = clientQueries.ReferSubmission(ctx, clientdb.ReferSubmissionParams{
			ID:          submissionID,
			ReviewedBy:  pgtype.UUID{Bytes: reviewedBy, Valid: true},
			ReviewNotes: req.RejectionNotes,
		})

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid action",
		})
	}

	if err != nil {
		h.logger.Errorw("Failed to review submission", "error", err, "action", req.Action)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to review submission",
		})
	}

	response := buildSubmissionResponse(submission)

	h.logger.Infow("Submission reviewed", 
		"submission_id", submissionID, 
		"action", req.Action, 
		"reviewer_id", reviewedBy,
		"client_id", clientID)

	return c.JSON(http.StatusOK, response)
}

// ListSubmissionsByStatus lists submissions filtered by status
func (h *Handler) ListSubmissionsByStatus(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	status := c.QueryParam("status")
	if status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Status query parameter is required",
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

	// List submissions
	submissions, err := clientQueries.ListSubmissionsByStatus(ctx, clientdb.SubmissionStatusEnum(status))
	if err != nil {
		h.logger.Errorw("Failed to list submissions", "error", err, "status", status)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve submissions",
		})
	}

	// Convert to response format
	responses := make([]SubmissionResponse, 0, len(submissions))
	for _, sub := range submissions {
		responses = append(responses, buildSubmissionResponse(clientdb.Submission{
			ID:              sub.ID,
			QuestionID:      sub.QuestionID,
			SubmittedBy:     sub.SubmittedBy,
			AnswerText:      sub.AnswerText,
			AnswerValue:     sub.AnswerValue,
			Explanation:     sub.Explanation,
			Status:          sub.Status,
			Version:         sub.Version,
			ReviewedBy:      sub.ReviewedBy,
			ReviewedAt:      sub.ReviewedAt,
			ReviewNotes:     sub.ReviewNotes,
			RejectionReason: sub.RejectionReason,
			SubmittedAt:     sub.SubmittedAt,
			CreatedAt:       sub.CreatedAt,
			UpdatedAt:       sub.UpdatedAt,
		}))
	}

	return c.JSON(http.StatusOK, responses)
}

// GetSubmission retrieves a specific submission
func (h *Handler) GetSubmission(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	submissionID, err := uuid.Parse(c.Param("submissionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
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

	// Get submission
	submission, err := clientQueries.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		h.logger.Errorw("Failed to get submission", "error", err, "submission_id", submissionID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Submission not found",
		})
	}

	response := buildSubmissionResponse(submission)

	return c.JSON(http.StatusOK, response)
}

// Helper function to build submission response
func buildSubmissionResponse(submission clientdb.Submission) SubmissionResponse {
	var answer *string
	if submission.AnswerText != nil {
		answer = submission.AnswerText
	}

	var answerValue *string
	if submission.AnswerValue.Valid {
		av := string(submission.AnswerValue.AnswerValueEnum)
		answerValue = &av
	}

	var reviewedBy *string
	if submission.ReviewedBy.Valid {
		rb := uuid.UUID(submission.ReviewedBy.Bytes).String()
		reviewedBy = &rb
	}

	var reviewedAt *string
	if submission.ReviewedAt.Valid {
		ra := submission.ReviewedAt.Time.Format(time.RFC3339)
		reviewedAt = &ra
	}

	var rejectionNotes *string
	if submission.ReviewNotes != nil {
		rejectionNotes = submission.ReviewNotes
	} else if submission.RejectionReason != nil {
		rejectionNotes = submission.RejectionReason
	}

	var submittedAt *string
	if submission.SubmittedAt.Valid {
		sa := submission.SubmittedAt.Time.Format(time.RFC3339)
		submittedAt = &sa
	}

	return SubmissionResponse{
		ID:             submission.ID.String(),
		QuestionID:     submission.QuestionID.String(),
		SubmittedBy:    submission.SubmittedBy.String(),
		Answer:         answer,
		AnswerValue:    answerValue,
		Status:         string(submission.Status),
		Version:        submission.Version,
		ReviewedBy:     reviewedBy,
		ReviewedAt:     reviewedAt,
		RejectionNotes: rejectionNotes,
		SubmittedAt:    submittedAt,
		CreatedAt:      submission.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      submission.UpdatedAt.Time.Format(time.RFC3339),
	}
}
