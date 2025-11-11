# Client Dashboard Implementation

## Overview
Implemented a client-specific dashboard that displays audit cycle enrollments, due dates, and framework analytics showing questions answered vs total questions for each framework.

## Changes Made

### 1. Backend - Tenant Service

#### SQL Queries Added

**File**: `/services/tenant-service/db/queries/dashboard.sql`

Added three new queries for client dashboard:

1. **GetClientAuditCycleEnrollments** - Retrieves all audit cycles a client is enrolled in with framework details
2. **CountClientActiveAuditCycles** - Counts active audit cycles for a specific client
3. **CountClientTotalFrameworkAssignments** - Counts total framework assignments across all audit cycles

**File**: `/services/tenant-service/db/client-queries/audits.sql`

Added query for framework analytics:

- **GetAllAuditsProgress** - Gets progress for all audits/frameworks with total questions and answered questions count

#### Dashboard Handler

**File**: `/services/tenant-service/internal/handler/dashboard.go`

Added new types and endpoint:

**Types**:
- `ClientDashboardStats` - Statistics for client dashboard
- `AuditCycleEnrollment` - Audit cycle enrollment with frameworks
- `FrameworkAssignment` - Framework assignment details
- `FrameworkAnalytics` - Analytics for framework progress
- `ClientDashboardData` - Complete client dashboard data

**Endpoint**:
- `GET /api/tenant/dashboard/client/:client_id` - Returns client-specific dashboard data

**Features**:
- Fetches audit cycle enrollments from tenant database
- Groups frameworks by audit cycle
- Connects to client-specific database to fetch framework analytics
- Returns questions answered vs total questions for each framework

#### Router

**File**: `/services/tenant-service/internal/router/router.go`

Added route:
```go
tenant.GET("/dashboard/client/:client_id", h.GetClientDashboard)
```

### 2. Frontend

#### Types

**File**: `/apps/frontend/app/types/index.ts`

Added new TypeScript interfaces:

```typescript
interface ClientDashboardStats {
  active_audit_cycles: number;
  total_framework_assignments: number;
}

interface FrameworkAssignment {
  framework_assignment_id?: string;
  framework_id?: string;
  framework_name?: string;
  due_date?: string;
  framework_status?: string;
  auditor_id?: string;
}

interface AuditCycleEnrollment {
  audit_cycle_id: string;
  audit_cycle_name: string;
  audit_cycle_description?: string;
  start_date: string;
  end_date: string;
  cycle_status: string;
  enrollment_id: string;
  enrolled_at: string;
  frameworks: FrameworkAssignment[];
}

interface FrameworkAnalytics {
  audit_id: string;
  framework_id: string;
  framework_name: string;
  status: string;
  due_date: string;
  total_questions: number;
  answered_questions: number;
}

interface ClientDashboardData {
  client_name: string;
  stats: ClientDashboardStats;
  audit_cycles: AuditCycleEnrollment[];
  framework_analytics: FrameworkAnalytics[];
}
```

#### API Client

**File**: `/apps/frontend/app/api/index.ts`

Added new API method:
```typescript
getClientSpecificDashboard: (clientId: string): Promise<AxiosResponse<ClientDashboardData>>
```

#### Dashboard Component

**File**: `/apps/frontend/app/routes/dashboard.tsx`

**Key Features**:

1. **User Detection**: Automatically detects if user is a client (has `client_id`)
2. **Conditional Rendering**: Shows different dashboard based on user type
3. **Client Dashboard Components**:
   - **Stats Cards**: Active audit cycles and total framework assignments
   - **Audit Cycle Enrollments**: Lists all audit cycles with:
     - Cycle name, description, start/end dates
     - Cycle status badge
     - Assigned frameworks with due dates and status
   - **Framework Analytics**: Shows for each framework:
     - Framework name and status
     - Due date
     - Progress percentage
     - Questions answered vs total questions
     - Visual progress bar

## Data Flow

```
1. User logs in
   ↓
2. Frontend checks if user.client_id exists
   ↓
3. If client user:
   - Call GET /api/tenant/dashboard/client/{client_id}
   ↓
4. Backend:
   - Fetch audit cycle enrollments from tenant DB
   - Group frameworks by audit cycle
   - Connect to client-specific DB
   - Fetch framework analytics (questions answered/total)
   ↓
5. Frontend:
   - Display stats cards
   - Show audit cycle enrollments with frameworks
   - Display framework analytics with progress bars
```

## API Endpoint

### Get Client Dashboard

```http
GET /api/tenant/dashboard/client/:client_id
Authorization: Cookie (auth_token)
```

**Response**:
```json
{
  "client_name": "Acme Corporation",
  "stats": {
    "active_audit_cycles": 2,
    "total_framework_assignments": 5
  },
  "audit_cycles": [
    {
      "audit_cycle_id": "uuid",
      "audit_cycle_name": "Q1 2024 Compliance Audit",
      "audit_cycle_description": "Quarterly compliance review",
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-03-31T00:00:00Z",
      "cycle_status": "active",
      "enrollment_id": "uuid",
      "enrolled_at": "2024-01-01T00:00:00Z",
      "frameworks": [
        {
          "framework_assignment_id": "uuid",
          "framework_id": "uuid",
          "framework_name": "SOC 2 Type II",
          "due_date": "2024-03-15T00:00:00Z",
          "framework_status": "in_progress",
          "auditor_id": "uuid"
        }
      ]
    }
  ],
  "framework_analytics": [
    {
      "audit_id": "uuid",
      "framework_id": "uuid",
      "framework_name": "SOC 2 Type II",
      "status": "in_progress",
      "due_date": "2024-03-15T00:00:00Z",
      "total_questions": 150,
      "answered_questions": 75
    }
  ]
}
```

## UI Features

### Client Dashboard View

1. **Header**
   - Dynamic title: "{Company Name} Dashboard" (e.g., "Acme Corporation Dashboard")
   - Personalized greeting with user name
   - Context-specific subtitle

2. **Stats Section**
   - Active Audit Cycles count
   - Total Framework Assignments count

3. **Audit Cycle Enrollments Card**
   - Lists all enrolled audit cycles
   - Shows cycle details (name, description, dates)
   - Status badges (active/completed/archived)
   - Nested framework list with:
     - Framework name
     - Due date
     - Status badge (pending/in_progress/completed/overdue)

4. **Framework Analytics Card**
   - Framework name and status
   - Due date display
   - Progress percentage (large, prominent)
   - Questions answered / total questions
   - Visual progress bar
   - Color-coded status badges

## Database Schema

### Tenant Database Tables Used

- `audit_cycles` - Audit cycle definitions
- `audit_cycle_clients` - Client enrollments in cycles
- `audit_cycle_frameworks` - Framework assignments to clients

### Client Database Tables Used

- `audits` - Framework assignments for the client
- `questions` - Questions for each framework
- `submissions` - Client answers/submissions

## Testing

### Manual Testing Steps

1. **Login as Client User**:
   - User must have `client_id` set
   - Navigate to `/dashboard`

2. **Verify Dashboard Display**:
   - Should see "Client Dashboard" header
   - Stats cards show correct counts
   - Audit cycles list appears with frameworks
   - Framework analytics show progress bars

3. **Check Data Accuracy**:
   - Verify audit cycle dates are correct
   - Confirm framework statuses match actual state
   - Validate question counts and progress percentages

### Test Endpoint

```bash
# Get client dashboard (replace with actual client_id)
curl -H "Cookie: auth_token=YOUR_TOKEN" \
  http://localhost:8080/api/tenant/dashboard/client/{client_id}
```

## Future Enhancements

1. **Filtering & Sorting**
   - Filter by audit cycle status
   - Sort frameworks by due date or progress
   - Search frameworks by name

2. **Interactive Features**
   - Click framework to view questions
   - Quick actions to start/continue frameworks
   - Export progress reports

3. **Notifications**
   - Overdue framework alerts
   - Upcoming due date reminders
   - Completion notifications

4. **Advanced Analytics**
   - Historical progress trends
   - Comparison across audit cycles
   - Time-to-completion metrics

## Summary

✅ **Backend**:
- SQL queries for audit cycle enrollments
- SQL queries for framework analytics
- Client dashboard handler with proper data grouping
- New API endpoint for client-specific dashboard

✅ **Frontend**:
- TypeScript types for client dashboard data
- API client method for fetching data
- Conditional dashboard rendering based on user type
- Beautiful UI with stats, audit cycles, and analytics
- Progress bars and status badges
- Responsive design

The client dashboard now provides a comprehensive view of audit cycle enrollments and framework completion progress, making it easy for clients to track their compliance obligations and progress.
