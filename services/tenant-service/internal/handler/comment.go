package handler

import (
	"net/http"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CommentRequest represents the request body for creating/updating comments
type CommentRequest struct {
	SubmissionID string `json:"submission_id"`
	CommentText  string `json:"comment_text" validate:"required"`
	IsInternal   bool   `json:"is_internal"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID           string `json:"id"`
	SubmissionID string `json:"submission_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	CommentText  string `json:"comment_text"`
	IsInternal   bool   `json:"is_internal"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreateComment creates a new comment on a submission
func (h *Handler) CreateComment(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	var req CommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.CommentText == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Comment text is required",
		})
	}

	submissionID, err := uuid.Parse(req.SubmissionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
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

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Create comment
	comment, err := clientQueries.CreateComment(ctx, clientdb.CreateCommentParams{
		SubmissionID: submissionID,
		UserID:       userUUID,
		UserName:     userEmail, // Using email as display name for now
		CommentText:  req.CommentText,
		IsInternal:   req.IsInternal,
	})
	if err != nil {
		h.logger.Errorw("Failed to create comment", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create comment",
		})
	}

	h.logger.Infow("Comment created", 
		"comment_id", comment.ID, 
		"submission_id", submissionID, 
		"client_id", clientID,
		"user_id", userID)

	response := buildCommentResponse(comment)

	return c.JSON(http.StatusCreated, response)
}

// GetComment retrieves a specific comment
func (h *Handler) GetComment(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid comment ID",
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

	// Get comment
	comment, err := clientQueries.GetCommentByID(ctx, commentID)
	if err != nil {
		h.logger.Errorw("Failed to get comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Comment not found",
		})
	}

	response := buildCommentResponse(comment)

	return c.JSON(http.StatusOK, response)
}

// ListCommentsBySubmission lists all comments for a submission
func (h *Handler) ListCommentsBySubmission(c echo.Context) error {
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

	// Get filter parameter (all, internal, external)
	filter := c.QueryParam("filter")
	if filter == "" {
		filter = "all"
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// List comments based on filter
	var comments []clientdb.Comment

	switch filter {
	case "internal":
		comments, err = clientQueries.ListInternalComments(ctx, submissionID)
	case "external":
		comments, err = clientQueries.ListExternalComments(ctx, submissionID)
	default:
		comments, err = clientQueries.ListCommentsBySubmission(ctx, submissionID)
	}

	if err != nil {
		h.logger.Errorw("Failed to list comments", "error", err, "submission_id", submissionID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve comments",
		})
	}

	// Convert to response format
	responses := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		responses = append(responses, buildCommentResponse(comment))
	}

	return c.JSON(http.StatusOK, responses)
}

// UpdateComment updates a comment's text
func (h *Handler) UpdateComment(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid comment ID",
		})
	}

	var req struct {
		CommentText string `json:"comment_text" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.CommentText == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Comment text is required",
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

	// Update comment
	comment, err := clientQueries.UpdateComment(ctx, clientdb.UpdateCommentParams{
		ID:          commentID,
		CommentText: req.CommentText,
	})
	if err != nil {
		h.logger.Errorw("Failed to update comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update comment",
		})
	}

	h.logger.Infow("Comment updated", "comment_id", commentID, "client_id", clientID)

	response := buildCommentResponse(comment)

	return c.JSON(http.StatusOK, response)
}

// DeleteComment deletes a comment
func (h *Handler) DeleteComment(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid comment ID",
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

	// Delete comment
	err = clientQueries.DeleteComment(ctx, commentID)
	if err != nil {
		h.logger.Errorw("Failed to delete comment", "error", err, "comment_id", commentID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete comment",
		})
	}

	h.logger.Infow("Comment deleted", "comment_id", commentID, "client_id", clientID)

	return c.NoContent(http.StatusNoContent)
}

// Helper function to build comment response
func buildCommentResponse(comment clientdb.Comment) CommentResponse {
	return CommentResponse{
		ID:           comment.ID.String(),
		SubmissionID: comment.SubmissionID.String(),
		UserID:       comment.UserID.String(),
		UserName:     comment.UserName,
		CommentText:  comment.CommentText,
		IsInternal:   comment.IsInternal,
		CreatedAt:    comment.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    comment.UpdatedAt.Time.Format(time.RFC3339),
	}
}
