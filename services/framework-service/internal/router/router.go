package router

import (
	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/packages/go/rbac"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/config"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/handler"
	"github.com/NormaTech-AI/audity/services/framework-service/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for the framework service
func SetupRoutes(e *echo.Echo, h *handler.Handler, cfg *config.Config, st *store.Store, log *zap.SugaredLogger) {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API v1 routes - all require authentication
	api := e.Group("/api")
	api.Use(auth.AuthMiddleware(cfg.Auth.JWTSecret, log))

	// Framework management routes (protected)
	frameworks := api.Group("/frameworks")
	{
		frameworks.POST("",
			h.CreateFramework,
			rbac.PermissionMiddleware(st, log, "frameworks:create"),
		)

		frameworks.PUT("/:id",
			h.UpdateFramework,
			rbac.PermissionMiddleware(st, log, "frameworks:update"),
		)

		frameworks.DELETE("/:id",
			h.DeleteFramework,
			rbac.PermissionMiddleware(st, log, "frameworks:delete"),
		)

		frameworks.GET("",
			h.ListFrameworks,
			rbac.PermissionMiddleware(st, log, "frameworks:list"),
		)

		frameworks.GET("/:id",
			h.GetFramework,
			rbac.PermissionMiddleware(st, log, "frameworks:read"),
		)

		frameworks.GET("/:id/checklist",
			h.GetFrameworkChecklist,
			rbac.PermissionMiddleware(st, log, "frameworks:read"),
		)
	}
}
