# Role to Designation Refactoring Summary

## Overview
Refactored the `users` table to rename the `role` column to `designation` to avoid confusion with the actual RBAC `user_roles` table. The `/validate` endpoint now returns both the user's designation (job title) and their actual RBAC roles with appropriate module visibility.

## Changes Made

### 1. Database Schema Changes

#### Auth Service - Migration Files
- **Updated**: `services/auth-service/db/migrations/000001_users_schema.sql`
  - Renamed `user_role_enum` to `user_designation_enum`
  - Renamed `role` column to `designation` in users table
  - Updated index from `idx_users_role` to `idx_users_designation`
  - Added comments clarifying designation vs RBAC roles

- **Created**: `services/auth-service/db/migrations/000002_clients_table.sql`
  - Added minimal clients table for auth-service queries

#### SQL Queries
- **Updated**: `services/auth-service/db/queries/users.sql`
  - Changed all `role` references to `designation`
  - Removed duplicate `GetClientByEmailDomain` query

- **Created**: `services/auth-service/db/queries/clients.sql`
  - Added `GetClientByEmailDomain` query for clients table

### 2. Generated Code (sqlc)
- Regenerated all database code in `services/auth-service/internal/db/`
- Updated enum from `UserRoleEnum` to `UserDesignationEnum`
- All constants updated (e.g., `UserDesignationEnumStakeholder`)

### 3. API Response Changes

#### UserInfo Structure
**Before:**
```go
type UserInfo struct {
    ID             uuid.UUID  `json:"id"`
    Email          string     `json:"email"`
    Name           string     `json:"name"`
    Role           string     `json:"role"`
    ClientID       *uuid.UUID `json:"client_id,omitempty"`
    VisibleModules []string   `json:"visible_modules"`
}
```

**After:**
```go
type UserInfo struct {
    ID             uuid.UUID  `json:"id"`
    Email          string     `json:"email"`
    Name           string     `json:"name"`
    Designation    string     `json:"designation"`     // Job title
    Roles          []string   `json:"roles"`           // RBAC roles
    ClientID       *uuid.UUID `json:"client_id,omitempty"`
    VisibleModules []string   `json:"visible_modules"`
}
```

### 4. Handler Logic Updates

#### New Function: `getUserRoles()`
- Fetches actual RBAC roles from the `user_roles` table in tenant-service
- Joins `roles` and `user_roles` tables
- Returns array of role names

#### Updated Function: `getVisibleModules()`
- Now takes both `roles []string` and `designation string` parameters
- Prioritizes RBAC roles over designation
- Admin roles (`admin`, `super_admin`) get full access
- Falls back to designation-based access if no roles assigned

#### Updated Functions:
- `findOrCreateUser()` - Now fetches roles and populates visible modules
- `getUserByID()` - Now fetches roles and populates visible modules
- `convertDBUserToUserInfo()` - Updated to use designation parameter

### 5. Module Visibility Logic

**Admin Role** (from user_roles):
- Dashboard, Clients, Users, Roles and Permission, Assessments

**Designation Fallback** (when no roles assigned):
- `nishaj_admin`: All modules
- `auditor`: Dashboard, Clients, Assessments
- `team_member`: Dashboard, Assessments
- `poc_internal`/`poc_client`: Dashboard, Assessments
- `stakeholder`: Dashboard, Assessments
- Default: Dashboard only

## API Response Example

### Admin User (with RBAC role)
```json
{
  "id": "uuid",
  "email": "admin@nishaj.com",
  "name": "Admin User",
  "designation": "nishaj_admin",
  "roles": ["admin", "user_manager"],
  "client_id": null,
  "visible_modules": ["Dashboard", "Clients", "Users", "Roles and Permission", "Assessments"]
}
```

### Stakeholder User (no RBAC roles yet)
```json
{
  "id": "uuid",
  "email": "stakeholder@client.com",
  "name": "Stakeholder User",
  "designation": "stakeholder",
  "roles": [],
  "client_id": "uuid",
  "visible_modules": ["Dashboard", "Assessments"]
}
```

## Migration Notes

### For Existing Databases
If you have an existing database with the old schema, you'll need to run a migration:

```sql
-- Rename the column
ALTER TABLE users RENAME COLUMN role TO designation;

-- Rename the index
DROP INDEX IF EXISTS idx_users_role;
CREATE INDEX idx_users_designation ON users(designation);

-- Rename the enum type
ALTER TYPE user_role_enum RENAME TO user_designation_enum;
```

### For New Deployments
The updated schema in `000001_users_schema.sql` will be used automatically.

## Frontend Integration

The frontend should now:
1. Use the `designation` field to display the user's job title
2. Use the `roles` array for permission checks
3. Use the `visible_modules` array to show/hide navigation items

Example frontend code:
```typescript
// Check if user has admin role
const isAdmin = user.roles.includes('admin') || user.roles.includes('super_admin');

// Show modules based on visible_modules
const shouldShowModule = (moduleName: string) => {
  return user.visible_modules.includes(moduleName);
};
```

## Testing Checklist

- [ ] Test `/validate` endpoint returns correct structure
- [ ] Verify users with RBAC roles get correct modules
- [ ] Verify users without RBAC roles fall back to designation
- [ ] Test new user creation (should have empty roles array)
- [ ] Verify admin users see all modules
- [ ] Verify stakeholders see only Dashboard and Assessments

## Files Modified

### Database
- `services/auth-service/db/migrations/000001_users_schema.sql`
- `services/auth-service/db/migrations/000002_clients_table.sql` (new)
- `services/auth-service/db/queries/users.sql`
- `services/auth-service/db/queries/clients.sql` (new)

### Generated Code
- `services/auth-service/internal/db/models.go`
- `services/auth-service/internal/db/users.sql.go`
- `services/auth-service/internal/db/clients.sql.go` (new)
- `services/auth-service/internal/db/querier.go`

### Handler Code
- `services/auth-service/internal/handler/auth.go`

### Configuration
- `services/auth-service/sqlc.yaml`
