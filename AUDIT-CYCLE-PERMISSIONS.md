# Audit Cycle Permissions by Role

## Overview
This document outlines the permissions assigned to each role for the Audit Cycle module.

## Permissions Defined

| Permission Name | Resource | Action | Description |
|----------------|----------|--------|-------------|
| `audit_cycles:create` | audit_cycles | create | Create audit cycles |
| `audit_cycles:read` | audit_cycles | read | View audit cycle details |
| `audit_cycles:update` | audit_cycles | update | Update audit cycles |
| `audit_cycles:delete` | audit_cycles | delete | Delete audit cycles |
| `audit_cycles:list` | audit_cycles | list | List all audit cycles |
| `audit_cycles:manage_clients` | audit_cycles | manage_clients | Add/remove clients from audit cycles |
| `audit_cycles:assign_frameworks` | audit_cycles | assign_frameworks | Assign frameworks to clients in audit cycles |

## Role Permissions Matrix

### 1. **nishaj_admin** (System Administrator)
**Full Access** - All audit cycle permissions

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ✅ |
| audit_cycles:read | ✅ |
| audit_cycles:update | ✅ |
| audit_cycles:delete | ✅ |
| audit_cycles:list | ✅ |
| audit_cycles:manage_clients | ✅ |
| audit_cycles:assign_frameworks | ✅ |

**Use Cases:**
- Create and manage audit cycles
- Add/remove clients from cycles
- Assign frameworks to clients
- Update cycle status and details
- Delete audit cycles

---

### 2. **auditor** (Primary Reviewer)
**Read-Only Access** - Can view audit cycles but not modify

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ❌ |
| audit_cycles:read | ✅ |
| audit_cycles:update | ❌ |
| audit_cycles:delete | ❌ |
| audit_cycles:list | ✅ |
| audit_cycles:manage_clients | ❌ |
| audit_cycles:assign_frameworks | ❌ |

**Use Cases:**
- View audit cycle details
- See which clients are in which cycles
- Review framework assignments
- Monitor audit cycle progress

---

### 3. **team_member** (Support Staff)
**Read-Only Access** - Can view audit cycles

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ❌ |
| audit_cycles:read | ✅ |
| audit_cycles:update | ❌ |
| audit_cycles:delete | ❌ |
| audit_cycles:list | ✅ |
| audit_cycles:manage_clients | ❌ |
| audit_cycles:assign_frameworks | ❌ |

**Use Cases:**
- View audit cycle information
- See client assignments
- Support auditors with cycle information

---

### 4. **poc_internal** (Internal Point of Contact)
**Read-Only Access** - Can view audit cycles

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ❌ |
| audit_cycles:read | ✅ |
| audit_cycles:update | ❌ |
| audit_cycles:delete | ❌ |
| audit_cycles:list | ✅ |
| audit_cycles:manage_clients | ❌ |
| audit_cycles:assign_frameworks | ❌ |

**Use Cases:**
- View audit cycle schedules
- See which clients are in active cycles
- Coordinate with clients about upcoming audits

---

### 5. **poc_client** (Client Point of Contact)
**No Access** - Cannot see audit cycle management

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ❌ |
| audit_cycles:read | ❌ |
| audit_cycles:update | ❌ |
| audit_cycles:delete | ❌ |
| audit_cycles:list | ❌ |
| audit_cycles:manage_clients | ❌ |
| audit_cycles:assign_frameworks | ❌ |

**Rationale:**
- Clients should not see internal audit cycle planning
- They only need to see their assigned audits/frameworks
- Keeps internal planning separate from client view

---

### 6. **stakeholder** (Client Employee)
**No Access** - Cannot see audit cycle management

| Permission | Granted |
|-----------|---------|
| audit_cycles:create | ❌ |
| audit_cycles:read | ❌ |
| audit_cycles:update | ❌ |
| audit_cycles:delete | ❌ |
| audit_cycles:list | ❌ |
| audit_cycles:manage_clients | ❌ |
| audit_cycles:assign_frameworks | ❌ |

**Rationale:**
- Stakeholders only need to answer assigned questions
- No need to see audit cycle management interface
- Focused on task completion, not planning

---

## Permission Usage in Routes

The following routes use these permissions:

```go
// List all audit cycles
GET /api/audit-cycles
Permission: audit_cycles:list

// Create audit cycle
POST /api/audit-cycles
Permission: audit_cycles:create

// Get specific audit cycle
GET /api/audit-cycles/:id
Permission: audit_cycles:read

// Update audit cycle
PUT /api/audit-cycles/:id
Permission: audit_cycles:update

// Delete audit cycle
DELETE /api/audit-cycles/:id
Permission: audit_cycles:delete

// Get audit cycle statistics
GET /api/audit-cycles/:id/stats
Permission: audit_cycles:read

// Get clients in audit cycle
GET /api/audit-cycles/:id/clients
Permission: audit_cycles:read

// Add client to audit cycle
POST /api/audit-cycles/:id/clients
Permission: audit_cycles:update (or audit_cycles:manage_clients)

// Remove client from audit cycle
DELETE /api/audit-cycles/:id/clients/:clientId
Permission: audit_cycles:update (or audit_cycles:manage_clients)

// Get frameworks in audit cycle
GET /api/audit-cycles/:id/frameworks
Permission: audit_cycles:read

// Assign framework to client
POST /api/audit-cycles/clients/:cycleClientId/frameworks
Permission: audit_cycles:update (or audit_cycles:assign_frameworks)
```

## Recommended Updates to Routes

Consider updating the router to use more specific permissions:

```go
// Current: uses audit_cycles:update
auditCycles.POST("/:id/clients",
    h.AddClientToAuditCycle,
    rbac.PermissionMiddleware(store, logger, "audit_cycles:manage_clients"), // More specific
)

// Current: uses audit_cycles:update
auditCycles.POST("/clients/:cycleClientId/frameworks",
    h.AssignFrameworkToClient,
    rbac.PermissionMiddleware(store, logger, "audit_cycles:assign_frameworks"), // More specific
)
```

## Migration Status

✅ Migration `000005_add_audit_cycle_permissions` successfully applied
✅ All permissions created in database
✅ Role permissions assigned correctly

## Testing Permissions

To verify permissions are working:

1. **As nishaj_admin**: Should be able to create, edit, delete audit cycles
2. **As auditor**: Should only be able to view audit cycles
3. **As team_member**: Should only be able to view audit cycles
4. **As poc_internal**: Should only be able to view audit cycles
5. **As poc_client**: Should NOT see audit cycles menu/pages
6. **As stakeholder**: Should NOT see audit cycles menu/pages

## Future Enhancements

Consider adding these permissions if needed:
- `audit_cycles:export` - Export audit cycle data
- `audit_cycles:archive` - Archive completed cycles
- `audit_cycles:report` - Generate cycle reports
- `audit_cycles:notify` - Send notifications about cycles
