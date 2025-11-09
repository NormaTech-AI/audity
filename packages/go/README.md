# Shared Go Packages

This directory contains shared Go packages used across all microservices in the Audity TPRM platform.

## Structure

```
packages/go/
├── auth/           # Authentication & JWT management
├── rbac/           # Role-Based Access Control middleware
├── config/         # (Future) Shared configuration utilities
├── logger/         # (Future) Shared logging setup
├── database/       # (Future) Database connection utilities
└── validator/      # (Future) Request validation utilities
```

## Packages

### 1. auth

**Purpose:** JWT token management and authentication middleware

**Features:**
- JWT token generation
- JWT token validation
- Token refresh
- Authentication middleware for Echo framework
- User claims management

**Usage:**
```go
import "github.com/NormaTech-AI/audity/packages/go/auth"

// Create JWT manager
jwtManager := auth.NewJWTManager(jwtSecret, expirationHours)

// Generate token
token, err := jwtManager.GenerateToken(userID, email, name, role, clientID)

// Validate token
claims, err := jwtManager.ValidateToken(tokenString)

// Use middleware in Echo
e.Use(auth.AuthMiddleware(jwtSecret, logger))

// Get user from context in handlers
user, err := auth.GetUserFromContext(c)
```

**Exports:**
- `JWTClaims` - Token claims structure
- `JWTManager` - Token management
- `AuthMiddleware` - Echo middleware for JWT validation
- `GetUserFromContext` - Helper to extract user from context

---

### 2. rbac

**Purpose:** Role-Based Access Control middleware

**Features:**
- Permission-based access control
- Role-based access control
- Database-backed permission checking
- Multiple permission strategies (AND/OR logic)

**Usage:**
```go
import "github.com/NormaTech-AI/audity/packages/go/rbac"

// Single permission check
router.POST("/clients", handler.CreateClient,
    rbac.PermissionMiddleware(store, logger, "clients:create"))

// Multiple permissions (AND logic)
router.PUT("/clients/:id", handler.UpdateClient,
    rbac.RequirePermissions(store, logger, "clients:update", "clients:read"))

// Any permission (OR logic)
router.GET("/clients", handler.ListClients,
    rbac.RequireAnyPermission(store, logger, "clients:list", "clients:read"))

// Role-based
router.GET("/admin", handler.AdminPanel,
    rbac.RequireRole(logger, "nishaj_admin"))

// Any role (OR logic)
router.GET("/internal", handler.InternalDashboard,
    rbac.RequireAnyRole(logger, "nishaj_admin", "auditor", "poc_internal"))
```

**Exports:**
- `PermissionMiddleware` - Single permission check
- `RequirePermissions` - Multiple permissions (AND)
- `RequireAnyPermission` - At least one permission (OR)
- `RequireRole` - Single role check
- `RequireAnyRole` - Multiple roles (OR)

**Store Interface:**
```go
type Store interface {
    GetPool() *pgxpool.Pool
}
```

---

## Go Workspace

This monorepo uses Go workspaces (`go.work`) to manage multiple modules:

```go
go 1.24.4

use (
    ./packages/go/auth
    ./packages/go/rbac
    ./services/auth-service
    ./services/tenant-service
)
```

### Benefits:
1. **Shared Code** - No duplication across services
2. **Type Safety** - Compile-time checks across modules
3. **Easy Refactoring** - Changes propagate automatically
4. **Version Control** - Single source of truth
5. **Development Speed** - Local changes reflect immediately

---

## Development

### Adding a New Shared Package

1. Create package directory:
```bash
mkdir -p packages/go/mypackage
```

2. Initialize module:
```bash
cd packages/go/mypackage
go mod init github.com/NormaTech-AI/audity/packages/go/mypackage
```

3. Add to workspace:
```bash
# Edit go.work at repository root
use (
    ...
    ./packages/go/mypackage
)
```

4. Use in services:
```go
// In service's go.mod
require (
    github.com/NormaTech-AI/audity/packages/go/mypackage v0.0.0
)

replace github.com/NormaTech-AI/audity/packages/go/mypackage => ../../packages/go/mypackage
```

### Using Shared Packages

In any service:

```go
// go.mod
require (
    github.com/NormaTech-AI/audity/packages/go/auth v0.0.0
    github.com/NormaTech-AI/audity/packages/go/rbac v0.0.0
)

replace (
    github.com/NormaTech-AI/audity/packages/go/auth => ../../packages/go/auth
    github.com/NormaTech-AI/audity/packages/go/rbac => ../../packages/go/rbac
)
```

```go
// main.go or handlers
import (
    "github.com/NormaTech-AI/audity/packages/go/auth"
    "github.com/NormaTech-AI/audity/packages/go/rbac"
)
```

### Running Tests

```bash
# Test all packages
go test ./packages/go/...

# Test specific package
go test ./packages/go/auth
go test ./packages/go/rbac

# With coverage
go test -cover ./packages/go/...
```

### Building Services

```bash
# Build all services
go build ./services/...

# Build specific service
cd services/auth-service
go build -o bin/auth-service main.go

cd services/tenant-service
go build -o bin/tenant-service main.go
```

---

## Best Practices

### 1. Keep Packages Focused
- Each package should have a single, clear responsibility
- Avoid circular dependencies
- Use interfaces for loose coupling

### 2. Version Management
- Use `v0.0.0` for internal packages
- Use `replace` directives for local development
- Consider semantic versioning for stable APIs

### 3. Documentation
- Document all exported functions and types
- Include usage examples
- Keep README files up to date

### 4. Testing
- Write unit tests for all packages
- Aim for >80% code coverage
- Test edge cases and error handling

### 5. Breaking Changes
- Avoid breaking changes when possible
- If necessary, version the package (v2, v3, etc.)
- Communicate changes to all service teams

---

## Migration from Service-Specific Code

### Before (Duplicated Code)
```
services/
├── auth-service/
│   └── internal/
│       ├── auth/jwt.go          # Duplicated
│       └── middleware/auth.go   # Duplicated
└── tenant-service/
    └── internal/
        ├── middleware/auth.go   # Duplicated
        └── middleware/permission.go
```

### After (Shared Packages)
```
packages/go/
├── auth/
│   ├── jwt.go          # Shared
│   └── middleware.go   # Shared
└── rbac/
    └── permission.go   # Shared

services/
├── auth-service/
│   └── internal/
│       └── oidc/       # Service-specific
└── tenant-service/
    └── internal/
        └── handler/    # Service-specific
```

### Benefits:
- ✅ No code duplication
- ✅ Single source of truth
- ✅ Easier maintenance
- ✅ Consistent behavior across services
- ✅ Faster development

---

## Future Packages

### config
Shared configuration utilities:
- Environment variable parsing
- Configuration validation
- Default values
- Type-safe config structs

### logger
Centralized logging setup:
- Zap logger configuration
- Log level management
- Structured logging helpers
- Request ID tracking

### database
Database utilities:
- Connection pool management
- Migration helpers
- Transaction utilities
- Query builders

### validator
Request validation:
- Custom validators
- Error formatting
- Validation rules
- Echo integration

---

## Troubleshooting

### "Package not found"
```bash
# Run from repository root
go work sync
go mod tidy
```

### "Circular dependency"
- Review package dependencies
- Use interfaces to break cycles
- Consider splitting packages

### "Version mismatch"
```bash
# Update all modules
cd packages/go/auth && go mod tidy
cd ../rbac && go mod tidy
cd ../../services/auth-service && go mod tidy
cd ../tenant-service && go mod tidy
```

### "Changes not reflecting"
```bash
# Rebuild from scratch
go clean -cache
go build ./...
```

---

## Contributing

1. **Create Package** - Follow naming conventions
2. **Write Tests** - Ensure good coverage
3. **Document** - Add godoc comments
4. **Update README** - Keep documentation current
5. **Test Services** - Verify all services still work
6. **Submit PR** - Get code review

---

## License

Proprietary - Nishaj Infotech
