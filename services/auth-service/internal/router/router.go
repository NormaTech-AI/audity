package router

import (
	"github.com/NormaTech-AI/audity/packages/go/auth"
	"github.com/NormaTech-AI/audity/services/auth-service/internal/handler"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(e *echo.Echo, h *handler.Handler, jwtSecret string, logger *zap.SugaredLogger) {
	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	
	// CORS middleware - allow requests from frontend
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Requested-With", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))
	
	e.Use(echomiddleware.RequestID())

	// Root and health endpoints
	e.GET("/", h.RootHandler)
	e.GET("/health", h.HealthCheck)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Auth routes (public)
	authGroup := e.Group("/auth")
	{
		// OAuth login initiation
		authGroup.GET("/login/:provider", h.InitiateLogin)
		
		// OAuth callback
		authGroup.GET("/callback", h.HandleCallback)
		
		// Set token as cookie (for OAuth flow)
		authGroup.POST("/set-token", h.SetTokenCookie)
		
		// Token refresh (requires valid token)
		authGroup.POST("/refresh", h.RefreshToken)
		
		// Logout
		authGroup.POST("/logout", h.Logout)
	}

	// Protected routes (require JWT)
	protected := e.Group("/auth")
	protected.Use(auth.AuthMiddleware(jwtSecret, logger))
	{
		// Validate token
		protected.GET("/validate", h.ValidateToken)
	}
}
