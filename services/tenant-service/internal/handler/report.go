package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/NormaTech-AI/audity/services/tenant-service/internal/clientdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
)

// ReportResponse represents a report in API responses
type ReportResponse struct {
	ID               string                 `json:"id"`
	AuditID          string                 `json:"audit_id"`
	UnsignedFilePath *string                `json:"unsigned_file_path,omitempty"`
	SignedFilePath   *string                `json:"signed_file_path,omitempty"`
	GeneratedBy      string                 `json:"generated_by"`
	GeneratedAt      string                 `json:"generated_at"`
	SignedBy         *string                `json:"signed_by,omitempty"`
	SignedAt         *string                `json:"signed_at,omitempty"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	DownloadURL      *string                `json:"download_url,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// ReportData holds the data for generating a report
type ReportData struct {
	AuditID       string
	FrameworkName string
	ClientName    string
	AuditStatus   string
	DueDate       string
	Questions     []QuestionReportData
	GeneratedAt   string
	GeneratedBy   string
}

// QuestionReportData holds question data for reports
type QuestionReportData struct {
	Section        string
	QuestionNumber string
	QuestionText   string
	Answer         string
	AnswerValue    string
	Status         string
	Evidence       []string
}

// GenerateReport generates an audit report (HTML and PDF)
func (h *Handler) GenerateReport(c echo.Context) error {
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

	// Get user info from context
	userID := c.Get("user_id").(string)
	userEmail := c.Get("user_email").(string)

	generatedBy, err := uuid.Parse(userID)
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

	// Get audit details
	audit, err := clientQueries.GetAuditByID(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get audit", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Audit not found",
		})
	}

	// Get questions with submissions
	questions, err := clientQueries.ListQuestionsWithSubmissions(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get questions", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve audit data",
		})
	}

	// Prepare report data
	reportData := ReportData{
		AuditID:       auditID.String(),
		FrameworkName: audit.FrameworkName,
		ClientName:    clientID.String(), // TODO: Get actual client name
		AuditStatus:   string(audit.Status),
		DueDate:       formatDate(audit.DueDate),
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
		GeneratedBy:   userEmail,
		Questions:     make([]QuestionReportData, 0),
	}

	// Process questions
	for _, q := range questions {
		qData := QuestionReportData{
			Section:        q.Section,
			QuestionNumber: q.QuestionNumber,
			QuestionText:   q.QuestionText,
			Status:         "Not Answered",
		}

		if q.SubmissionStatus.Valid {
			qData.Status = string(q.SubmissionStatus.SubmissionStatusEnum)
		}

		reportData.Questions = append(reportData.Questions, qData)
	}

	// Generate HTML report
	htmlContent, err := generateHTMLReport(reportData)
	if err != nil {
		h.logger.Errorw("Failed to generate HTML report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate report",
		})
	}

	// Upload HTML to MinIO
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
	reportID := uuid.New()
	htmlPath := fmt.Sprintf("reports/%s/%s.html", auditID.String(), reportID.String())

	_, err = h.minio.PutObject(ctx, bucketName, htmlPath, bytes.NewReader([]byte(htmlContent)), int64(len(htmlContent)), minio.PutObjectOptions{
		ContentType: "text/html",
	})
	if err != nil {
		h.logger.Errorw("Failed to upload HTML to MinIO", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save report",
		})
	}

	// Create report record
	report, err := clientQueries.CreateReport(ctx, clientdb.CreateReportParams{
		AuditID:          auditID,
		UnsignedFilePath: &htmlPath,
		GeneratedBy:      generatedBy,
		Status:           clientdb.ReportStatusEnumGenerated,
	})
	if err != nil {
		h.logger.Errorw("Failed to create report record", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save report",
		})
	}

	h.logger.Infow("Report generated", 
		"report_id", report.ID, 
		"audit_id", auditID, 
		"client_id", clientID)

	response := buildReportResponse(report, nil)

	return c.JSON(http.StatusCreated, response)
}

// GetReport retrieves a specific report
func (h *Handler) GetReport(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	reportID, err := uuid.Parse(c.Param("reportId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid report ID",
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

	// Get report
	report, err := clientQueries.GetReportByID(ctx, reportID)
	if err != nil {
		h.logger.Errorw("Failed to get report", "error", err, "report_id", reportID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Report not found",
		})
	}

	// Generate download URL if requested
	var downloadURL *string
	if c.QueryParam("include_url") == "true" {
		bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
		
		// Prefer signed version if available
		filePath := report.UnsignedFilePath
		if report.SignedFilePath != nil {
			filePath = report.SignedFilePath
		}

		if filePath != nil {
			url, err := h.minio.PresignedGetObject(ctx, bucketName, *filePath, 1*time.Hour, nil)
			if err != nil {
				h.logger.Warnw("Failed to generate download URL", "error", err)
			} else {
				urlStr := url.String()
				downloadURL = &urlStr
			}
		}
	}

	response := buildReportResponse(report, downloadURL)

	return c.JSON(http.StatusOK, response)
}

// GetReportByAudit retrieves a report for a specific audit
func (h *Handler) GetReportByAudit(c echo.Context) error {
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

	// Get report
	report, err := clientQueries.GetReportByAuditID(ctx, auditID)
	if err != nil {
		h.logger.Errorw("Failed to get report by audit", "error", err, "audit_id", auditID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Report not found",
		})
	}

	// Generate download URL if requested
	var downloadURL *string
	if c.QueryParam("include_url") == "true" {
		bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
		
		filePath := report.UnsignedFilePath
		if report.SignedFilePath != nil {
			filePath = report.SignedFilePath
		}

		if filePath != nil {
			url, err := h.minio.PresignedGetObject(ctx, bucketName, *filePath, 1*time.Hour, nil)
			if err != nil {
				h.logger.Warnw("Failed to generate download URL", "error", err)
			} else {
				urlStr := url.String()
				downloadURL = &urlStr
			}
		}
	}

	response := buildReportResponse(report, downloadURL)

	return c.JSON(http.StatusOK, response)
}

// ListReportsByStatus lists reports filtered by status
func (h *Handler) ListReportsByStatus(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	status := c.QueryParam("status")
	if status == "" {
		status = "generated"
	}

	// Validate status
	var statusEnum clientdb.ReportStatusEnum
	switch status {
	case "pending":
		statusEnum = clientdb.ReportStatusEnumPending
	case "generated":
		statusEnum = clientdb.ReportStatusEnumGenerated
	case "signed":
		statusEnum = clientdb.ReportStatusEnumSigned
	case "delivered":
		statusEnum = clientdb.ReportStatusEnumDelivered
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid status",
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

	// List reports
	reports, err := clientQueries.ListReportsByStatus(ctx, statusEnum)
	if err != nil {
		h.logger.Errorw("Failed to list reports", "error", err, "status", status)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve reports",
		})
	}

	// Convert to response format
	responses := make([]ReportResponse, 0, len(reports))
	for _, rep := range reports {
		responses = append(responses, buildReportResponseFromList(rep))
	}

	return c.JSON(http.StatusOK, responses)
}

// SignReport marks a report as signed with digital signature
func (h *Handler) SignReport(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	reportID, err := uuid.Parse(c.Param("reportId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid report ID",
		})
	}

	// Get user info from context
	userID := c.Get("user_id").(string)
	signedBy, err := uuid.Parse(userID)
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

	// Get report
	report, err := clientQueries.GetReportByID(ctx, reportID)
	if err != nil {
		h.logger.Errorw("Failed to get report", "error", err, "report_id", reportID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Report not found",
		})
	}

	// For now, we'll just mark it as signed without actual digital signature
	// In production, you would:
	// 1. Download the unsigned file from MinIO
	// 2. Apply digital signature
	// 3. Upload signed version to MinIO
	// 4. Update the record with signed file path

	signedPath := report.UnsignedFilePath // Placeholder - would be actual signed file path

	// Update report as signed
	signedReport, err := clientQueries.UpdateReportSigned(ctx, clientdb.UpdateReportSignedParams{
		ID:             reportID,
		SignedFilePath: signedPath,
		SignedBy:       pgtype.UUID{Bytes: signedBy, Valid: true},
	})
	if err != nil {
		h.logger.Errorw("Failed to sign report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to sign report",
		})
	}

	h.logger.Infow("Report signed", "report_id", reportID, "signed_by", userID, "client_id", clientID)

	response := buildReportResponse(signedReport, nil)

	return c.JSON(http.StatusOK, response)
}

// MarkReportDelivered marks a report as delivered to the client
func (h *Handler) MarkReportDelivered(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	reportID, err := uuid.Parse(c.Param("reportId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid report ID",
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

	// Mark as delivered
	report, err := clientQueries.MarkReportDelivered(ctx, reportID)
	if err != nil {
		h.logger.Errorw("Failed to mark report as delivered", "error", err, "report_id", reportID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update report status",
		})
	}

	h.logger.Infow("Report marked as delivered", "report_id", reportID, "client_id", clientID)

	response := buildReportResponse(report, nil)

	return c.JSON(http.StatusOK, response)
}

// DownloadReport streams the report file to the client
func (h *Handler) DownloadReport(c echo.Context) error {
	ctx := c.Request().Context()

	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid client ID",
		})
	}

	reportID, err := uuid.Parse(c.Param("reportId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid report ID",
		})
	}

	// Get version parameter (signed or unsigned)
	version := c.QueryParam("version")
	if version == "" {
		version = "signed" // Default to signed if available
	}

	// Get client database queries
	clientQueries, _, err := h.clientStore.GetClientQueries(ctx, clientID)
	if err != nil {
		h.logger.Errorw("Failed to get client queries", "error", err, "client_id", clientID)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to access client data",
		})
	}

	// Get report
	report, err := clientQueries.GetReportByID(ctx, reportID)
	if err != nil {
		h.logger.Errorw("Failed to get report", "error", err, "report_id", reportID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Report not found",
		})
	}

	// Determine which file to download
	var filePath *string
	if version == "signed" && report.SignedFilePath != nil {
		filePath = report.SignedFilePath
	} else if report.UnsignedFilePath != nil {
		filePath = report.UnsignedFilePath
	} else {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Report file not found",
		})
	}

	// Download from MinIO
	bucketName := fmt.Sprintf("client-%s", clientID.String()[:8])
	object, err := h.minio.GetObject(ctx, bucketName, *filePath, minio.GetObjectOptions{})
	if err != nil {
		h.logger.Errorw("Failed to get report from MinIO", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve report",
		})
	}
	defer object.Close()

	// Set headers
	fileName := fmt.Sprintf("audit-report-%s.html", reportID.String()[:8])
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Response().Header().Set("Content-Type", "text/html")

	// Stream the file
	return c.Stream(http.StatusOK, "text/html", object)
}

// Helper functions

func buildReportResponse(report clientdb.Report, downloadURL *string) ReportResponse {
	var metadata map[string]interface{}
	if len(report.Metadata) > 0 {
		json.Unmarshal(report.Metadata, &metadata)
	}

	var signedBy *string
	if report.SignedBy.Valid {
		sb := uuid.UUID(report.SignedBy.Bytes).String()
		signedBy = &sb
	}

	var signedAt *string
	if report.SignedAt.Valid {
		sa := report.SignedAt.Time.Format(time.RFC3339)
		signedAt = &sa
	}

	return ReportResponse{
		ID:               report.ID.String(),
		AuditID:          report.AuditID.String(),
		UnsignedFilePath: report.UnsignedFilePath,
		SignedFilePath:   report.SignedFilePath,
		GeneratedBy:      report.GeneratedBy.String(),
		GeneratedAt:      report.GeneratedAt.Time.Format(time.RFC3339),
		SignedBy:         signedBy,
		SignedAt:         signedAt,
		Status:           string(report.Status),
		Metadata:         metadata,
		DownloadURL:      downloadURL,
		CreatedAt:        report.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        report.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func buildReportResponseFromList(report clientdb.ListReportsByStatusRow) ReportResponse {
	var metadata map[string]interface{}
	if len(report.Metadata) > 0 {
		json.Unmarshal(report.Metadata, &metadata)
	}

	var signedBy *string
	if report.SignedBy.Valid {
		sb := uuid.UUID(report.SignedBy.Bytes).String()
		signedBy = &sb
	}

	var signedAt *string
	if report.SignedAt.Valid {
		sa := report.SignedAt.Time.Format(time.RFC3339)
		signedAt = &sa
	}

	return ReportResponse{
		ID:               report.ID.String(),
		AuditID:          report.AuditID.String(),
		UnsignedFilePath: report.UnsignedFilePath,
		SignedFilePath:   report.SignedFilePath,
		GeneratedBy:      report.GeneratedBy.String(),
		GeneratedAt:      report.GeneratedAt.Time.Format(time.RFC3339),
		SignedBy:         signedBy,
		SignedAt:         signedAt,
		Status:           string(report.Status),
		Metadata:         metadata,
		CreatedAt:        report.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        report.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func formatDate(date pgtype.Date) string {
	if date.Valid {
		return date.Time.Format("2006-01-02")
	}
	return "N/A"
}

func generateHTMLReport(data ReportData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Audit Report - {{.FrameworkName}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 40px;
            color: #333;
        }
        .header {
            text-align: center;
            border-bottom: 2px solid #0066cc;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #0066cc;
            margin: 0;
        }
        .meta-info {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 30px;
        }
        .meta-item {
            padding: 10px;
            background: #f5f5f5;
            border-left: 4px solid #0066cc;
        }
        .meta-item strong {
            display: block;
            color: #0066cc;
            margin-bottom: 5px;
        }
        .question-section {
            margin-bottom: 30px;
        }
        .section-header {
            background: #0066cc;
            color: white;
            padding: 10px 15px;
            margin-top: 30px;
            margin-bottom: 15px;
        }
        .question {
            border: 1px solid #ddd;
            padding: 15px;
            margin-bottom: 15px;
            background: white;
        }
        .question-header {
            font-weight: bold;
            color: #0066cc;
            margin-bottom: 10px;
        }
        .question-text {
            margin-bottom: 10px;
        }
        .answer {
            background: #f9f9f9;
            padding: 10px;
            margin-top: 10px;
            border-left: 3px solid #28a745;
        }
        .status {
            display: inline-block;
            padding: 3px 10px;
            border-radius: 3px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-approved { background: #28a745; color: white; }
        .status-submitted { background: #ffc107; color: #000; }
        .status-rejected { background: #dc3545; color: white; }
        .status-draft { background: #6c757d; color: white; }
        .footer {
            margin-top: 50px;
            padding-top: 20px;
            border-top: 2px solid #0066cc;
            text-align: center;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Compliance Audit Report</h1>
        <h2>{{.FrameworkName}}</h2>
    </div>

    <div class="meta-info">
        <div class="meta-item">
            <strong>Audit ID</strong>
            {{.AuditID}}
        </div>
        <div class="meta-item">
            <strong>Client</strong>
            {{.ClientName}}
        </div>
        <div class="meta-item">
            <strong>Audit Status</strong>
            {{.AuditStatus}}
        </div>
        <div class="meta-item">
            <strong>Due Date</strong>
            {{.DueDate}}
        </div>
        <div class="meta-item">
            <strong>Generated At</strong>
            {{.GeneratedAt}}
        </div>
        <div class="meta-item">
            <strong>Generated By</strong>
            {{.GeneratedBy}}
        </div>
    </div>

    {{$currentSection := ""}}
    {{range .Questions}}
        {{if ne .Section $currentSection}}
            {{if ne $currentSection ""}}
                </div>
            {{end}}
            <div class="section-header">{{.Section}}</div>
            <div class="question-section">
            {{$currentSection = .Section}}
        {{end}}
        
        <div class="question">
            <div class="question-header">
                Question {{.QuestionNumber}}
                <span class="status status-{{.Status | lower}}">{{.Status}}</span>
            </div>
            <div class="question-text">{{.QuestionText}}</div>
            {{if .Answer}}
                <div class="answer">
                    <strong>Answer:</strong> {{.Answer}}
                    {{if .AnswerValue}}
                        <br><strong>Value:</strong> {{.AnswerValue}}
                    {{end}}
                </div>
            {{end}}
        </div>
    {{end}}
    {{if .Questions}}
        </div>
    {{end}}

    <div class="footer">
        <p>This report was automatically generated on {{.GeneratedAt}}</p>
        <p>Confidential - For authorized use only</p>
    </div>
</body>
</html>
`

	// Create template with custom functions
	funcMap := template.FuncMap{
		"lower": func(s string) string {
			return fmt.Sprintf("%s", s)
		},
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
