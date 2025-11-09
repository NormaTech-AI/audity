package router

import (
	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/services/client-service/internal/config"
	"github.com/NormaTech-AI/audity/services/client-service/internal/handler"
	"github.com/NormaTech-AI/audity/packages/go/rbac"
	"github.com/NormaTech-AI/audity/services/client-service/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(e *echo.Echo, h *handler.Handler, cfg *config.Config, store *store.Store, logger *zap.SugaredLogger) {
	// Global middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

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

	// API routes - all require authentication
	api := e.Group("/api")
	api.Use(auth.AuthMiddleware(cfg.Auth.JWTSecret, logger))

	// Dashboard routes (protected)
	client := api.Group("/client")
	{
		client.GET("/dashboard", h.GetClientDashboard)
		client.GET("/dashboard/stats", h.GetClientDashboardStats)
	}

	// Client management routes (protected)
	clients := api.Group("/clients")
	{
		// Admin only - create clients
		clients.POST("",
			h.CreateClient,
			rbac.PermissionMiddleware(store, logger, "clients:create"),
		)

		// Admin and internal POC - list clients
		clients.GET("",
			h.ListClients,
			rbac.RequireAnyPermission(store, logger, "clients:list", "clients:read"),
		)

		// Anyone authenticated can view client details (filtered by their access)
		clients.GET("/:id",
			h.GetClient,
			rbac.PermissionMiddleware(store, logger, "clients:read"),
		)

		// TODO: Add update and delete endpoints with proper permissions
		clients.PUT("/:id", h.UpdateClient, rbac.PermissionMiddleware(store, logger, "clients:update"))
		// clients.DELETE("/:id", h.DeleteClient, rbac.PermissionMiddleware(store, logger, "clients:delete"))
	}
}
