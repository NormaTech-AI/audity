package router

import (
	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/packages/go/rbac"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/tenant-service/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(e *echo.Echo, h *handler.Handler, cfg *config.Config, store *store.Store, logger *zap.SugaredLogger) {
	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())

	// CORS middleware - allow requests from frontend
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Requested-With", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	e.Use(middleware.RequestID())

	// Public endpoints (no auth required)
	e.GET("/", h.RootHandler)
	e.GET("/health", h.HealthCheck)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API routes - all require authentication
	api := e.Group("/api")
	api.Use(auth.AuthMiddleware(cfg.Auth.JWTSecret, logger))

	// Dashboard routes (protected)
	tenant := api.Group("/tenant")
	{
		tenant.GET("/dashboard", h.GetTenantDashboard)
		tenant.GET("/dashboard/stats", h.GetTenantDashboardStats)
		tenant.GET("/dashboard/client/:client_id", h.GetClientDashboard)
	}

	// Audit management routes (protected, client-specific)
	audits := api.Group("/clients/:clientId/audits")
	{
		// List audits for a client
		audits.GET("",
			h.ListClientAudits,
			rbac.RequireAnyPermission(store, logger, "audits:list", "audits:read"),
		)

		// Get specific audit with questions
		audits.GET("/:auditId",
			h.GetAudit,
			rbac.PermissionMiddleware(store, logger, "audits:read"),
		)

		// Update audit (assignment, status, due date)
		audits.PATCH("/:auditId",
			h.UpdateAudit,
			rbac.PermissionMiddleware(store, logger, "audits:update"),
		)
	}

	// Submission management routes (protected, client-specific)
	submissions := api.Group("/clients/:clientId/submissions")
	{
		// Create or update draft submission
		submissions.POST("",
			h.CreateOrUpdateSubmission,
			rbac.PermissionMiddleware(store, logger, "submissions:create"),
		)

		// Submit for review
		submissions.POST("/:submissionId/submit",
			h.SubmitForReview,
			rbac.PermissionMiddleware(store, logger, "submissions:submit"),
		)

		// Review submission (approve/reject/refer)
		submissions.POST("/:submissionId/review",
			h.ReviewSubmission,
			rbac.PermissionMiddleware(store, logger, "submissions:review"),
		)

		// List submissions by status
		submissions.GET("",
			h.ListSubmissionsByStatus,
			rbac.PermissionMiddleware(store, logger, "submissions:list"),
		)

		// Get specific submission
		submissions.GET("/:submissionId",
			h.GetSubmission,
			rbac.PermissionMiddleware(store, logger, "submissions:read"),
		)
	}

	// Evidence management routes (protected, client-specific)
	evidence := api.Group("/clients/:clientId/evidence")
	{
		// Upload evidence file
		evidence.POST("/upload",
			h.UploadEvidence,
			rbac.PermissionMiddleware(store, logger, "evidence:upload"),
		)

		// Get presigned upload URL
		evidence.GET("/upload-url",
			h.GetPresignedUploadURL,
			rbac.PermissionMiddleware(store, logger, "evidence:upload"),
		)

		// List evidence by submission
		evidence.GET("/submissions/:submissionId",
			h.ListEvidenceBySubmission,
			rbac.PermissionMiddleware(store, logger, "evidence:list"),
		)

		// Get specific evidence with download URL
		evidence.GET("/:evidenceId",
			h.GetEvidence,
			rbac.PermissionMiddleware(store, logger, "evidence:read"),
		)

		// Download evidence file directly
		evidence.GET("/:evidenceId/download",
			h.DownloadEvidence,
			rbac.PermissionMiddleware(store, logger, "evidence:read"),
		)

		// Delete evidence (soft delete)
		evidence.DELETE("/:evidenceId",
			h.DeleteEvidence,
			rbac.PermissionMiddleware(store, logger, "evidence:delete"),
		)
	}

	// Comment management routes (protected, client-specific)
	comments := api.Group("/clients/:clientId/comments")
	{
		// Create comment on submission
		comments.POST("",
			h.CreateComment,
			rbac.PermissionMiddleware(store, logger, "comments:create"),
		)

		// List comments by submission
		comments.GET("/submissions/:submissionId",
			h.ListCommentsBySubmission,
			rbac.PermissionMiddleware(store, logger, "comments:list"),
		)

		// Get specific comment
		comments.GET("/:commentId",
			h.GetComment,
			rbac.PermissionMiddleware(store, logger, "comments:read"),
		)

		// Update comment
		comments.PUT("/:commentId",
			h.UpdateComment,
			rbac.PermissionMiddleware(store, logger, "comments:update"),
		)

		// Delete comment
		comments.DELETE("/:commentId",
			h.DeleteComment,
			rbac.PermissionMiddleware(store, logger, "comments:delete"),
		)
	}

	// Activity log routes (protected, client-specific)
	activity := api.Group("/clients/:clientId/activity")
	{
		// Create activity log entry
		activity.POST("",
			h.CreateActivityLog,
			rbac.PermissionMiddleware(store, logger, "activity:create"),
		)

		// List all activity logs with pagination
		activity.GET("",
			h.ListActivityLogs,
			rbac.PermissionMiddleware(store, logger, "activity:list"),
		)

		// Get recent activity
		activity.GET("/recent",
			h.GetRecentActivity,
			rbac.PermissionMiddleware(store, logger, "activity:list"),
		)

		// List activity by user
		activity.GET("/users/:userId",
			h.ListActivityLogsByUser,
			rbac.PermissionMiddleware(store, logger, "activity:list"),
		)

		// List activity by entity
		activity.GET("/entities",
			h.ListActivityLogsByEntity,
			rbac.PermissionMiddleware(store, logger, "activity:list"),
		)
	}

	// Report generation routes (protected, client-specific)
	reports := api.Group("/clients/:clientId/reports")
	{
		// Generate new report for audit
		reports.POST("/audits/:auditId/generate",
			h.GenerateReport,
			rbac.PermissionMiddleware(store, logger, "reports:create"),
		)

		// Get report by ID
		reports.GET("/:reportId",
			h.GetReport,
			rbac.PermissionMiddleware(store, logger, "reports:read"),
		)

		// Get report by audit ID
		reports.GET("/audits/:auditId",
			h.GetReportByAudit,
			rbac.PermissionMiddleware(store, logger, "reports:read"),
		)

		// List reports by status
		reports.GET("",
			h.ListReportsByStatus,
			rbac.PermissionMiddleware(store, logger, "reports:list"),
		)

		// Sign report
		reports.POST("/:reportId/sign",
			h.SignReport,
			rbac.PermissionMiddleware(store, logger, "reports:sign"),
		)

		// Mark report as delivered
		reports.POST("/:reportId/deliver",
			h.MarkReportDelivered,
			rbac.PermissionMiddleware(store, logger, "reports:deliver"),
		)

		// Download report file
		reports.GET("/:reportId/download",
			h.DownloadReport,
			rbac.PermissionMiddleware(store, logger, "reports:read"),
		)
	}

	// Audit Cycle management routes (protected)
	auditCycles := api.Group("/audit-cycles")
	{
		// List all audit cycles
		auditCycles.GET("",
			h.ListAuditCycles,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:list"),
		)

		// Create audit cycle
		auditCycles.POST("",
			h.CreateAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:create"),
		)

		// Get specific audit cycle
		auditCycles.GET("/:id",
			h.GetAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:read"),
		)

		// Update audit cycle
		auditCycles.PUT("/:id",
			h.UpdateAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:update"),
		)

		// Delete audit cycle
		auditCycles.DELETE("/:id",
			h.DeleteAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:delete"),
		)

		// Get audit cycle statistics
		auditCycles.GET("/:id/stats",
			h.GetAuditCycleStats,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:read"),
		)

		// Client management within audit cycle
		auditCycles.GET("/:id/clients",
			h.GetAuditCycleClients,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:read"),
		)

		auditCycles.POST("/:id/clients",
			h.AddClientToAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:manage_clients"),
		)

		auditCycles.DELETE("/:id/clients/:clientId",
			h.RemoveClientFromAuditCycle,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:manage_clients"),
		)

		// Framework management within audit cycle
		auditCycles.GET("/:id/frameworks",
			h.GetAuditCycleFrameworks,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:read"),
		)

		auditCycles.POST("/clients/:cycleClientId/frameworks",
			h.AssignFrameworkToClient,
			rbac.PermissionMiddleware(store, logger, "audit_cycles:assign_frameworks"),
		)
	}

	// Users management routes (protected)
	users := api.Group("/users")
	{
		// List all users (with optional client_id filter)
		users.GET("",
			h.ListUsers,
			rbac.PermissionMiddleware(store, logger, "users:list"),
		)

		// Get specific user
		users.GET("/:id",
			h.GetUser,
			rbac.PermissionMiddleware(store, logger, "users:read"),
		)
	}

	// Client Audit View routes (for client users to view and submit)
	clientAudit := api.Group("/client-audit")
	{
		// List all audits for authenticated client
		clientAudit.GET("",
			h.ListClientAuditsView,
			rbac.PermissionMiddleware(store, logger, "audit:list"),
		)

		// Get audit detail with questions (role-based filtering)
		clientAudit.GET("/:auditId",
			h.GetClientAuditDetailView,
			rbac.PermissionMiddleware(store, logger, "audit:read"),
		)

		// Save submission (create or update draft)
		clientAudit.POST("/submissions",
			h.SaveClientSubmission,
			rbac.PermissionMiddleware(store, logger, "audit:submit"),
		)

		// Submit answer for review
		clientAudit.POST("/submissions/:submissionId/submit",
			h.SubmitClientAnswer,
			rbac.PermissionMiddleware(store, logger, "audit:submit"),
		)
	}
}
