# Client-Specific RBAC Implementation

This document describes the implementation of client-specific user roles and permissions using a global connection cache.

## Overview

Each client now has their own user roles and permissions stored in their isolated client database (`client_db`). The system uses a global connection cache to efficiently manage database connections across all client databases.

## Architecture

### 1. Database Schema

#### Client Database Schema (per-client)
Each client database now includes:
- `client_users` - Maps tenant users to client-specific roles
- `client_roles` - Client-specific role definitions
- `client_permissions` - Client-specific permission definitions
- `client_role_permissions` - Maps roles to permissions
- `client_user_roles` - Maps users to roles (many-to-many)

**Migrations:**
- `000002_add_client_user_roles.up.sql` - Creates the schema
- `000003_seed_client_roles_permissions.up.sql` - Seeds default roles and permissions

**Default Roles:**
- `client_admin` - Full administrative access
- `poc` - Point of Contact - can manage audits and delegate
- `stakeholder` - Can answer assigned questions
- `viewer` - Read-only access

### 2. Global Connection Cache

**Location:** `/packages/go/database/client_pool.go`

The `ClientPoolCache` manages connection pools for all client databases:
- Thread-safe with RWMutex
- Automatic connection health checking
- Lazy loading (connections created on first use)
- Automatic reconnection on failure

**Key Methods:**
```go
// Get or create a connection pool for a client
GetClientPool(ctx context.Context, clientID uuid.UUID) (*pgxpool.Pool, error)

// Remove a specific client's pool
RemovePool(clientID uuid.UUID)

// Close all pools
Close()
```

### 3. Enhanced Auth Middleware

**Location:** `/packages/go/auth/middleware.go`

Two middleware options:

#### Option 1: Basic Auth (Existing)
```go
AuthMiddleware(jwtSecret string, logger *zap.SugaredLogger)
```
- Validates JWT token
- Adds user context

#### Option 2: Auth with Client DB (New)
```go
AuthMiddlewareWithClientDB(
    jwtSecret string, 
    tenantStore TenantStore, 
    clientPoolCache ClientPoolCache, 
    logger *zap.SugaredLogger
)
```
- Validates JWT token
- Fetches user's client_id from tenant_db
- Gets client_db pool from cache
- Adds to context: `user`, `client_id`, `client_db`

**Helper Functions:**
```go
GetUserFromContext(c echo.Context) (*JWTClaims, error)
GetClientDBFromContext(c echo.Context) (*pgxpool.Pool, error)
GetClientIDFromContext(c echo.Context) (uuid.UUID, error)
```

### 4. Client-Specific RBAC

**Location:** `/packages/go/rbac/permission.go`

New middleware for client-specific permissions:

```go
// Check single permission in client_db
ClientPermissionMiddleware(logger *zap.SugaredLogger, requiredPermission string)

// Check any of multiple permissions (OR logic)
ClientRequireAnyPermission(logger *zap.SugaredLogger, permissions ...string)
```

**Helper Functions:**
```go
// Check if user has permission in client_db
checkClientUserPermission(ctx context.Context, clientDB *pgxpool.Pool, tenantUserID uuid.UUID, permissionName string) (bool, error)

// Get user's role from client_db
GetClientUserRole(ctx context.Context, clientDB *pgxpool.Pool, tenantUserID uuid.UUID) (string, error)
```

## Usage Example

### Step 1: Initialize Connection Cache

```go
import (
    "github.com/NormaTech-AI/audity/packages/go/database"
    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/NormaTech-AI/audity/packages/go/rbac"
)

// In main.go or service initialization
func main() {
    // ... initialize tenant database pool ...
    
    // Create decryptor (implement the Decryptor interface)
    decryptor := crypto.NewEncryptor(encryptionKey)
    
    // Create global client pool cache
    clientPoolCache := database.NewClientPoolCache(tenantPool, decryptor, logger)
    defer clientPoolCache.Close()
    
    // ... continue with server setup ...
}
```

### Step 2: Use Enhanced Auth Middleware

```go
// In router setup
func SetupRoutes(e *echo.Echo, tenantStore *store.Store, clientPoolCache *database.ClientPoolCache, cfg *config.Config, logger *zap.SugaredLogger) {
    // Public routes
    e.GET("/health", healthHandler)
    
    // API routes with client-specific auth
    api := e.Group("/api")
    api.Use(auth.AuthMiddlewareWithClientDB(
        cfg.Auth.JWTSecret,
        tenantStore,
        clientPoolCache,
        logger,
    ))
    
    // Protected routes with client-specific permissions
    audits := api.Group("/audits")
    {
        // Only users with 'audits:read' permission in their client_db
        audits.GET("", 
            listAuditsHandler,
            rbac.ClientPermissionMiddleware(logger, "audits:read"),
        )
        
        // Only users with 'audits:manage' permission
        audits.POST("", 
            createAuditHandler,
            rbac.ClientPermissionMiddleware(logger, "audits:manage"),
        )
        
        // Users with either permission
        audits.GET("/:id", 
            getAuditHandler,
            rbac.ClientRequireAnyPermission(logger, "audits:read", "audits:manage"),
        )
    }
}
```

### Step 3: Use Client DB in Handlers

```go
func listAuditsHandler(c echo.Context) error {
    // Get client_db from context
    clientDB, err := auth.GetClientDBFromContext(c)
    if err != nil {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "Client database access required",
        })
    }
    
    // Get user info
    user, _ := auth.GetUserFromContext(c)
    clientID, _ := auth.GetClientIDFromContext(c)
    
    // Query client-specific data
    query := `SELECT id, framework_name, status, due_date FROM audits ORDER BY created_at DESC`
    rows, err := clientDB.Query(c.Request().Context(), query)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to fetch audits",
        })
    }
    defer rows.Close()
    
    // ... process results ...
    
    return c.JSON(http.StatusOK, audits)
}
```

### Step 4: Add Users to Client Database

When a user is assigned to a client, they need to be added to that client's database:

```go
func addUserToClientDB(ctx context.Context, clientDB *pgxpool.Pool, tenantUserID uuid.UUID, email, name, role string) error {
    query := `
        INSERT INTO client_users (tenant_user_id, email, name, role, is_active)
        VALUES ($1, $2, $3, $4, true)
        ON CONFLICT (tenant_user_id) DO UPDATE
        SET email = EXCLUDED.email, name = EXCLUDED.name, role = EXCLUDED.role, is_active = true
    `
    
    _, err := clientDB.Exec(ctx, query, tenantUserID, email, name, role)
    return err
}
```

## Migration Guide

### For Existing Services

1. **Update imports:**
   ```go
   import (
       "github.com/NormaTech-AI/audity/packages/go/database"
       "github.com/NormaTech-AI/audity/packages/go/auth"
       "github.com/NormaTech-AI/audity/packages/go/rbac"
   )
   ```

2. **Initialize client pool cache in main.go:**
   ```go
   clientPoolCache := database.NewClientPoolCache(tenantPool, decryptor, logger)
   defer clientPoolCache.Close()
   ```

3. **Update middleware:**
   ```go
   // Old
   api.Use(auth.AuthMiddleware(cfg.Auth.JWTSecret, logger))
   
   // New
   api.Use(auth.AuthMiddlewareWithClientDB(cfg.Auth.JWTSecret, tenantStore, clientPoolCache, logger))
   ```

4. **Update RBAC middleware:**
   ```go
   // Old (tenant-level permissions)
   rbac.PermissionMiddleware(store, logger, "audits:read")
   
   // New (client-specific permissions)
   rbac.ClientPermissionMiddleware(logger, "audits:read")
   ```

5. **Run client database migrations:**
   ```bash
   # For each existing client database
   migrate -path services/tenant-service/db/client-migrations \
           -database "postgres://user:pass@host:5432/client_db_name" up
   ```

6. **Sync existing users to client databases:**
   Create a migration script to copy users from tenant_db to their respective client_db.

## Benefits

1. **Isolation:** Each client has complete control over their user roles
2. **Flexibility:** Clients can define custom roles and permissions
3. **Performance:** Connection pooling and caching minimize overhead
4. **Security:** Client data remains isolated in separate databases
5. **Scalability:** Global cache efficiently manages connections across all clients

## Monitoring

Track connection pool metrics:
```go
poolCount := clientPoolCache.GetPoolCount()
logger.Infow("Active client connections", "count", poolCount)
```

## Troubleshooting

### Connection Issues
- Check `client_databases` table in tenant_db for correct credentials
- Verify encryption key is correct
- Check database connectivity and firewall rules

### Permission Denied
- Verify user exists in `client_users` table
- Check user's role assignments in `client_user_roles`
- Verify role has required permissions in `client_role_permissions`

### Performance Issues
- Monitor connection pool sizes
- Adjust `MaxConns` and `MinConns` in `ClientPoolCache`
- Consider connection pool timeouts
