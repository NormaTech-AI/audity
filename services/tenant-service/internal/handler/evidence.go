package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
)

// EvidenceResponse represents evidence in API responses
type EvidenceResponse struct {
	ID           string  `json:"id"`
	SubmissionID string  `json:"submission_id"`
	FileName     string  `json:"file_name"`
	FileType     string  `json:"file_type"`
	FileSize     int64   `json:"file_size"`
	StoragePath  string  `json:"storage_path"`
	UploadedBy   string  `json:"uploaded_by"`
	Description  *string `json:"description"`
	DownloadURL  *string `json:"download_url,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// UploadEvidenceResponse includes upload URL for presigned uploads
type UploadEvidenceResponse struct {
	EvidenceID  string `json:"evidence_id"`
	UploadURL   string `json:"upload_url"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

const (
	maxFileSize        = 50 * 1024 * 1024 // 50MB
	presignedURLExpiry = 15 * time.Minute
	downloadURLExpiry  = 1 * time.Hour
)

var allowedFileTypes = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".txt":  true,
	".csv":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".zip":  true,
	".rar":  true,
}

// UploadEvidence handles direct file upload
func (h *Handler) UploadEvidence(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	submissionID, err := uuid.Parse(c.FormValue("submission_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
		})
	}

	description := c.FormValue("description")

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "File is required",
		})
	}

	// Validate file size
	if file.Size > maxFileSize {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("File size exceeds maximum allowed size of %dMB", maxFileSize/(1024*1024)),
		})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedFileTypes[ext] {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("File type %s is not allowed", ext),
		})
	}

	// Get user ID from context
	userID := c.Get("user_id").(string)
	uploadedBy, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
		})
	}

	// Open file for reading
	src, err := file.Open()
	if err != nil {
		h.logger.Errorw("Failed to open uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process file",
		})
	}
	defer src.Close()

	// Generate unique file path
	evidenceID := uuid.New()
	objectName := fmt.Sprintf("submissions/%s/%s%s", submissionID.String(), evidenceID.String(), ext)
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])

	// Upload to MinIO
	_, err = h.minio.PutObject(ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		h.logger.Errorw("Failed to upload to MinIO", "error", err, "bucket", bucketName, "object", objectName)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to upload file",
		})
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		// Try to delete the uploaded file
		h.minio.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Create evidence record
	var desc *string
	if description != "" {
		desc = &description
	}

	evidence, err := clientQueries.CreateEvidence(ctx, clientdb.CreateEvidenceParams{
		SubmissionID: submissionID,
		FileName:     file.Filename,
		FilePath:     objectName,
		FileSize:     file.Size,
		FileType:     &ext,
		UploadedBy:   uploadedBy,
		Description:  desc,
	})
	if err != nil {
		h.logger.Errorw("Failed to create evidence record", "error", err)
		// Try to delete the uploaded file
		h.minio.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create evidence record",
		})
	}

	h.logger.Infow("Evidence uploaded", 
		"evidence_id", evidenceID, 
		"submission_id", submissionID, 
		"client_id", clientID,
		"file_name", file.Filename)

	response := buildEvidenceResponse(evidence, nil)

	return c.JSON(http.StatusCreated, response)
}

// GetPresignedUploadURL generates a presigned URL for direct upload to MinIO
func (h *Handler) GetPresignedUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	submissionID, err := uuid.Parse(c.QueryParam("submission_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid submission ID",
		})
	}

	fileName := c.QueryParam("file_name")
	if fileName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "File name is required",
		})
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedFileTypes[ext] {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("File type %s is not allowed", ext),
		})
	}

	// Generate unique file path
	evidenceID := uuid.New()
	objectName := fmt.Sprintf("submissions/%s/%s%s", submissionID.String(), evidenceID.String(), ext)
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])

	// Generate presigned URL
	presignedURL, err := h.minio.PresignedPutObject(ctx, bucketName, objectName, presignedURLExpiry)
	if err != nil {
		h.logger.Errorw("Failed to generate presigned URL", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate upload URL",
		})
	}

	response := UploadEvidenceResponse{
		EvidenceID:  evidenceID.String(),
		UploadURL:   presignedURL.String(),
		ExpiresIn:   int(presignedURLExpiry.Seconds()),
	}

	h.logger.Infow("Presigned upload URL generated", 
		"evidence_id", evidenceID, 
		"submission_id", submissionID, 
		"client_id", clientID)

	return c.JSON(http.StatusOK, response)
}

// ListEvidenceBySubmission lists all evidence for a submission
func (h *Handler) ListEvidenceBySubmission(c echo.Context) error {
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

	// List evidence
	evidenceList, err := clientQueries.ListEvidenceBySubmission(ctx, submissionID)
	if err != nil {
		h.logger.Errorw("Failed to list evidence", "error", err, "submission_id", submissionID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve evidence",
		})
	}

	// Generate download URLs
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
	responses := make([]EvidenceResponse, 0, len(evidenceList))
	
	for _, ev := range evidenceList {
		var downloadURL *string
		if c.QueryParam("include_urls") == "true" {
			url, err := h.minio.PresignedGetObject(ctx, bucketName, ev.FilePath, downloadURLExpiry, nil)
			if err != nil {
				h.logger.Warnw("Failed to generate download URL", "error", err, "evidence_id", ev.ID)
			} else {
				urlStr := url.String()
				downloadURL = &urlStr
			}
		}

		responses = append(responses, buildEvidenceResponse(ev, downloadURL))
	}

	return c.JSON(http.StatusOK, responses)
}

// GetEvidence retrieves a specific evidence record with download URL
func (h *Handler) GetEvidence(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	evidenceID, err := uuid.Parse(c.Param("evidenceId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid evidence ID",
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

	// Get evidence
	evidence, err := clientQueries.GetEvidenceByID(ctx, evidenceID)
	if err != nil {
		h.logger.Errorw("Failed to get evidence", "error", err, "evidence_id", evidenceID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Evidence not found",
		})
	}

	// Generate download URL
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
	downloadURL, err := h.minio.PresignedGetObject(ctx, bucketName, evidence.FilePath, downloadURLExpiry, nil)
	if err != nil {
		h.logger.Errorw("Failed to generate download URL", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate download URL",
		})
	}

	urlStr := downloadURL.String()
	response := buildEvidenceResponse(evidence, &urlStr)

	return c.JSON(http.StatusOK, response)
}

// DeleteEvidence soft deletes an evidence record
func (h *Handler) DeleteEvidence(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	evidenceID, err := uuid.Parse(c.Param("evidenceId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid evidence ID",
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

	// Get user ID from context
	userID := c.Get("user_id").(string)
	deletedBy, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid user ID",
		})
	}

	// Soft delete evidence
	_, err = clientQueries.SoftDeleteEvidence(ctx, clientdb.SoftDeleteEvidenceParams{
		ID:        evidenceID,
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
	if err != nil {
		h.logger.Errorw("Failed to delete evidence", "error", err, "evidence_id", evidenceID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete evidence",
		})
	}

	h.logger.Infow("Evidence deleted", "evidence_id", evidenceID, "client_id", clientID)

	return c.NoContent(http.StatusNoContent)
}

// DownloadEvidence streams a file directly to the client
func (h *Handler) DownloadEvidence(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	evidenceID, err := uuid.Parse(c.Param("evidenceId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid evidence ID",
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

	// Get evidence
	evidence, err := clientQueries.GetEvidenceByID(ctx, evidenceID)
	if err != nil {
		h.logger.Errorw("Failed to get evidence", "error", err, "evidence_id", evidenceID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Evidence not found",
		})
	}

	// Get object from MinIO
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
	object, err := h.minio.GetObject(ctx, bucketName, evidence.FilePath, minio.GetObjectOptions{})
	if err != nil {
		h.logger.Errorw("Failed to get object from MinIO", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve file",
		})
	}
	defer object.Close()

	// Set headers for download
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", evidence.FileName))
	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", evidence.FileSize))

	// Stream the file
	_, err = io.Copy(c.Response().Writer, object)
	if err != nil {
		h.logger.Errorw("Failed to stream file", "error", err)
		return err
	}

	h.logger.Infow("Evidence downloaded", "evidence_id", evidenceID, "client_id", clientID)

	return nil
}

// Helper function to build evidence response
func buildEvidenceResponse(evidence clientdb.Evidence, downloadURL *string) EvidenceResponse {
	var desc *string
	if evidence.Description != nil {
		desc = evidence.Description
	}

	var fileType string
	if evidence.FileType != nil {
		fileType = *evidence.FileType
	}

	return EvidenceResponse{
		ID:           evidence.ID.String(),
		SubmissionID: evidence.SubmissionID.String(),
		FileName:     evidence.FileName,
		FileType:     fileType,
		FileSize:     evidence.FileSize,
		StoragePath:  evidence.FilePath,
		UploadedBy:   evidence.UploadedBy.String(),
		Description:  desc,
		DownloadURL:  downloadURL,
		CreatedAt:    evidence.UploadedAt.Time.Format(time.RFC3339),
		UpdatedAt:    evidence.UploadedAt.Time.Format(time.RFC3339), // Using uploaded_at as updated_at
	}
}
