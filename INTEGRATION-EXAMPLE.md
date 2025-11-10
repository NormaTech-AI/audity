# Integration Example: Tenant Service

This document shows how to integrate the client-specific RBAC system into the tenant-service.

## Step 1: Update main.go

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/NormaTech-AI/audity/packages/go/database"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/crypto"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/handler"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/router"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/store"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/labstack/echo/v4"
    "go.uber.org/zap"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    log := logger.Sugar()

    // Load configuration
    cfg, err := config.LoadConfig(".")
    if err != nil {
        log.Fatalw("Failed to load config", "error", err)
    }

    // Validate configuration
    if err := cfg.Validate(); err != nil {
        log.Fatalw("Invalid configuration", "error", err)
    }

    // Initialize tenant database connection
    ctx := context.Background()
    tenantPool, err := pgxpool.New(ctx, cfg.Database.TenantDBURL)
    if err != nil {
        log.Fatalw("Failed to connect to tenant database", "error", err)
    }
    defer tenantPool.Close()

    if err := tenantPool.Ping(ctx); err != nil {
        log.Fatalw("Failed to ping tenant database", "error", err)
    }
    log.Info("Connected to tenant database")

    // Initialize store
    tenantStore := store.NewStore(tenantPool)

    // Initialize encryptor/decryptor
    encryptor := crypto.NewEncryptor(cfg.Encryption.Key)

    // Initialize global client pool cache
    clientPoolCache := database.NewClientPoolCache(tenantPool, encryptor, log)
    defer clientPoolCache.Close()
    log.Info("Initialized client pool cache")

    // Initialize handler
    h := handler.NewHandler(tenantStore, cfg, encryptor, log)

    // Initialize Echo server
    e := echo.New()
    e.HideBanner = true

    // Setup routes with client pool cache
    router.SetupRoutes(e, h, cfg, tenantStore, clientPoolCache, log)

    // Start server
    go func() {
        addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
        log.Infow("Starting server", "address", addr)
        if err := e.Start(addr); err != nil {
            log.Infow("Server stopped", "error", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := e.Shutdown(ctx); err != nil {
        log.Errorw("Server forced to shutdown", "error", err)
    }

    log.Info("Server exited")
}
```

## Step 2: Update router.go

```go
package router

import (
    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/NormaTech-AI/audity/packages/go/database"
    "github.com/NormaTech-AI/audity/packages/go/rbac"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/config"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/handler"
    "github.com/NormaTech-AI/audity/services/tenant-service/internal/store"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "go.uber.org/zap"
)

func SetupRoutes(
    e *echo.Echo,
    h *handler.Handler,
    cfg *config.Config,
    tenantStore *store.Store,
    clientPoolCache *database.ClientPoolCache,
    logger *zap.SugaredLogger,
) {
    // Global middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Public routes
    e.GET("/", h.RootHandler)
    e.GET("/health", h.HealthCheck)

    // API routes with enhanced auth middleware
    api := e.Group("/api")
    api.Use(auth.AuthMiddlewareWithClientDB(
        cfg.Auth.JWTSecret,
        tenantStore,
        clientPoolCache,
        logger,
    ))

    // Client-specific audit routes
    audits := api.Group("/audits")
    {
        // List audits - requires read permission in client_db
        audits.GET("",
            h.ListAudits,
            rbac.ClientPermissionMiddleware(logger, "audits:read"),
        )

        // Create audit - requires manage permission
        audits.POST("",
            h.CreateAudit,
            rbac.ClientPermissionMiddleware(logger, "audits:manage"),
        )

        // Get audit details
        audits.GET("/:id",
            h.GetAudit,
            rbac.ClientRequireAnyPermission(logger, "audits:read", "audits:manage"),
        )

        // Update audit
        audits.PUT("/:id",
            h.UpdateAudit,
            rbac.ClientPermissionMiddleware(logger, "audits:manage"),
        )
    }

    // Submission routes
    submissions := api.Group("/submissions")
    {
        // List submissions
        submissions.GET("",
            h.ListSubmissions,
            rbac.ClientPermissionMiddleware(logger, "submissions:read"),
        )

        // Create submission
        submissions.POST("",
            h.CreateSubmission,
            rbac.ClientPermissionMiddleware(logger, "submissions:create"),
        )

        // Update own submission
        submissions.PUT("/:id",
            h.UpdateSubmission,
            rbac.ClientRequireAnyPermission(logger, "submissions:update", "submissions:manage"),
        )
    }

    // Evidence routes
    evidence := api.Group("/evidence")
    {
        evidence.GET("",
            h.ListEvidence,
            rbac.ClientPermissionMiddleware(logger, "evidence:read"),
        )

        evidence.POST("",
            h.UploadEvidence,
            rbac.ClientPermissionMiddleware(logger, "evidence:upload"),
        )

        evidence.DELETE("/:id",
            h.DeleteEvidence,
            rbac.ClientPermissionMiddleware(logger, "evidence:delete"),
        )
    }

    // Admin routes (tenant-level permissions)
    admin := api.Group("/admin")
    admin.Use(rbac.RequireRole(tenantStore, logger, "nishaj_admin"))
    {
        admin.GET("/clients", h.ListAllClients)
        admin.POST("/clients/:id/users", h.AddUserToClient)
    }
}
```

## Step 3: Update Handler to Use Client DB

```go
package handler

import (
    "net/http"

    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/google/uuid"
    "github.com/labstack/echo/v4"
)

func (h *Handler) ListAudits(c echo.Context) error {
    // Get client_db from context
    clientDB, err := auth.GetClientDBFromContext(c)
    if err != nil {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "Client database access required",
        })
    }

    // Get user and client info
    user, _ := auth.GetUserFromContext(c)
    clientID, _ := auth.GetClientIDFromContext(c)

    h.logger.Infow("Listing audits",
        "user_id", user.UserID,
        "client_id", clientID)

    // Query client-specific audits
    query := `
        SELECT id, framework_id, framework_name, status, due_date, created_at
        FROM audits
        ORDER BY created_at DESC
    `

    rows, err := clientDB.Query(c.Request().Context(), query)
    if err != nil {
        h.logger.Errorw("Failed to query audits", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to fetch audits",
        })
    }
    defer rows.Close()

    var audits []Audit
    for rows.Next() {
        var audit Audit
        err := rows.Scan(
            &audit.ID,
            &audit.FrameworkID,
            &audit.FrameworkName,
            &audit.Status,
            &audit.DueDate,
            &audit.CreatedAt,
        )
        if err != nil {
            h.logger.Errorw("Failed to scan audit", "error", err)
            continue
        }
        audits = append(audits, audit)
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "audits": audits,
        "count":  len(audits),
    })
}

func (h *Handler) CreateAudit(c echo.Context) error {
    // Get client_db from context
    clientDB, err := auth.GetClientDBFromContext(c)
    if err != nil {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "Client database access required",
        })
    }

    user, _ := auth.GetUserFromContext(c)
    clientID, _ := auth.GetClientIDFromContext(c)

    var req CreateAuditRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }

    // Insert into client database
    query := `
        INSERT INTO audits (framework_id, framework_name, assigned_by, due_date, status)
        VALUES ($1, $2, $3, $4, 'not_started')
        RETURNING id, created_at
    `

    var auditID uuid.UUID
    var createdAt time.Time
    err = clientDB.QueryRow(
        c.Request().Context(),
        query,
        req.FrameworkID,
        req.FrameworkName,
        user.UserID,
        req.DueDate,
    ).Scan(&auditID, &createdAt)

    if err != nil {
        h.logger.Errorw("Failed to create audit", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to create audit",
        })
    }

    h.logger.Infow("Audit created",
        "audit_id", auditID,
        "user_id", user.UserID,
        "client_id", clientID)

    return c.JSON(http.StatusCreated, map[string]interface{}{
        "id":         auditID,
        "created_at": createdAt,
    })
}
```

## Step 4: Add User Management Handler

```go
// AddUserToClient adds a user to a client's database with specified role
func (h *Handler) AddUserToClient(c echo.Context) error {
    clientIDStr := c.Param("id")
    clientID, err := uuid.Parse(clientIDStr)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid client ID",
        })
    }

    var req AddUserToClientRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }

    // Get client database pool
    ctx := c.Request().Context()
    clientDB, err := h.clientPoolCache.GetClientPool(ctx, clientID)
    if err != nil {
        h.logger.Errorw("Failed to get client database", "error", err, "client_id", clientID)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to access client database",
        })
    }

    // Add user to client database
    query := `
        INSERT INTO client_users (tenant_user_id, email, name, role, is_active)
        VALUES ($1, $2, $3, $4, true)
        ON CONFLICT (tenant_user_id) DO UPDATE
        SET email = EXCLUDED.email, name = EXCLUDED.name, role = EXCLUDED.role, is_active = true
        RETURNING id
    `

    var userID uuid.UUID
    err = clientDB.QueryRow(ctx, query, req.UserID, req.Email, req.Name, req.Role).Scan(&userID)
    if err != nil {
        h.logger.Errorw("Failed to add user to client", "error", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to add user to client",
        })
    }

    h.logger.Infow("User added to client",
        "user_id", req.UserID,
        "client_id", clientID,
        "role", req.Role)

    return c.JSON(http.StatusCreated, map[string]interface{}{
        "id":      userID,
        "message": "User added to client successfully",
    })
}
```

## Step 5: Run Migrations

```bash
# Run migrations on all existing client databases
# You'll need to create a script to iterate through all client databases

#!/bin/bash

# Get list of all client databases from tenant_db
psql $TENANT_DB_URL -t -c "SELECT db_name FROM client_databases" | while read -r db_name; do
    if [ ! -z "$db_name" ]; then
        echo "Migrating $db_name..."
        migrate -path services/tenant-service/db/client-migrations \
                -database "postgres://user:pass@host:5432/$db_name" up
    fi
done
```

## Step 6: Sync Existing Users

Create a script to sync existing users from tenant_db to their client_db:

```sql
-- Run this for each client database
-- Replace $CLIENT_ID with actual client ID

INSERT INTO client_users (tenant_user_id, email, name, role, is_active)
SELECT 
    u.id,
    u.email,
    u.name,
    CASE 
        WHEN u.role = 'poc_client' THEN 'poc'::client_user_role_enum
        WHEN u.role = 'stakeholder' THEN 'stakeholder'::client_user_role_enum
        ELSE 'viewer'::client_user_role_enum
    END,
    true
FROM tenant_db.users u
WHERE u.client_id = '$CLIENT_ID'
ON CONFLICT (tenant_user_id) DO NOTHING;
```

## Testing

1. **Test connection cache:**
   ```bash
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8083/api/audits
   ```

2. **Test permissions:**
   ```bash
   # Should succeed for users with audits:read
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8083/api/audits
   
   # Should fail for users without audits:manage
   curl -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"framework_id":"...","due_date":"2024-12-31"}' \
        http://localhost:8083/api/audits
   ```

3. **Monitor connection pool:**
   Add endpoint to check pool status:
   ```go
   e.GET("/debug/pools", func(c echo.Context) error {
       count := clientPoolCache.GetPoolCount()
       return c.JSON(http.StatusOK, map[string]interface{}{
           "active_pools": count,
       })
   })
   ```
