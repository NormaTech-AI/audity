# Framework Service Permissions

This document describes the permission-based access control (RBAC) implemented in the Framework Service.

## Overview

The Framework Service uses role-based access control (RBAC) to manage access to framework operations. All API endpoints require authentication via JWT token, and specific operations require additional permissions.

## Authentication Flow

1. **JWT Authentication**: All `/api/v1/*` routes require a valid JWT token (via `Authorization: Bearer <token>` header or `auth_token` cookie)
2. **Permission Check**: After authentication, each endpoint checks if the user has the required permission(s)

## Framework Permissions

The following permissions control access to framework operations:

### Read Permissions

- **`frameworks:list`** - List all compliance frameworks
  - **Endpoint**: `GET /api/v1/frameworks`
  - **Description**: View a list of all available frameworks
  
- **`frameworks:read`** - View framework details and checklist
  - **Endpoints**: 
    - `GET /api/v1/frameworks/:id`
    - `GET /api/v1/frameworks/:id/checklist`
  - **Description**: View detailed information about a specific framework, including its full checklist

### Write Permissions

- **`frameworks:create`** - Create new frameworks
  - **Endpoint**: `POST /api/v1/frameworks`
  - **Description**: Create a new compliance framework with checklist
  - **Typical Roles**: Admin, Framework Manager

- **`frameworks:update`** - Update existing frameworks
  - **Endpoint**: `PUT /api/v1/frameworks/:id`
  - **Description**: Modify framework details, checklist, or version
  - **Typical Roles**: Admin, Framework Manager

- **`frameworks:delete`** - Delete frameworks
  - **Endpoint**: `DELETE /api/v1/frameworks/:id`
  - **Description**: Remove a framework from the system
  - **Typical Roles**: Admin

## Permission Hierarchy

Typical role assignments:

### Admin Role
- All framework permissions (create, read, update, delete, list)

### Framework Manager Role
- `frameworks:create`
- `frameworks:read`
- `frameworks:update`
- `frameworks:list`

### Auditor Role
- `frameworks:read`
- `frameworks:list`

### Client User Role
- `frameworks:read`
- `frameworks:list`

## Public Endpoints

The following endpoints do NOT require authentication:

- `GET /health` - Health check endpoint
- `GET /swagger/*` - API documentation

## Implementation Details

### Middleware Stack

Each protected endpoint uses the following middleware chain:

1. **AuthMiddleware** - Validates JWT token and extracts user claims
2. **PermissionMiddleware** - Checks if user has required permission(s)

Example from router:
```go
frameworks.POST("",
    h.CreateFramework,
    rbac.PermissionMiddleware(st, log, "frameworks:create"),
)
```

### Permission Storage

Permissions are stored in the database and checked via the RBAC package:

- **Tables**: `permissions`, `roles`, `role_permissions`, `user_roles`
- **Query**: Joins user roles with role permissions to determine access

### Error Responses

- **401 Unauthorized**: Missing or invalid JWT token
- **403 Forbidden**: Valid token but insufficient permissions
- **500 Internal Server Error**: Permission check failed due to system error

## Testing Permissions

To test permissions:

1. Obtain a JWT token via the auth service
2. Include token in request: `Authorization: Bearer <token>`
3. Ensure user has appropriate role with required permissions

## Adding New Permissions

To add new framework-related permissions:

1. Add permission to database via migration
2. Assign permission to appropriate roles
3. Add permission check to route in `internal/router/router.go`
4. Update this documentation

## Related Documentation

- [Auth Package](../../packages/go/auth/README.md) - JWT authentication
- [RBAC Package](../../packages/go/rbac/README.md) - Permission middleware
- [Database Schema](./db/migrations/) - Permission tables
