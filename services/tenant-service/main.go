package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	swagger "github.com/swaggo/echo-swagger"

	// This blank import is required for Swag to find the docs
	_ "github.com/NormaTech-AI/audity/services/tenant-service/docs"
)

// @title Tenant Service API
// @version 1.0
// @description Handles client onboarding and authentication.
// @host localhost:8080       // This will be the gateway's address
// @BasePath /api/tenant     // This will be the gateway's path
func main() {
	e := echo.New()

	// Health check
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"service": "tenant-service", "status": "ok"})
	})

	// Swagger docs route
	e.GET("/swagger/*", swagger.WrapHandler)

	// Run the service. It will NOT be exposed to the host.
	// The API Gateway will connect to this port.
	e.Logger.Fatal(e.Start(":8081")) // Running on internal port 8081
}