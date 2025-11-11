# Audit Cycle API Endpoints

## Base URL
```
http://localhost:8080/api
```

## Authentication
All endpoints require authentication via JWT token in cookie or Authorization header.

---

## Endpoints

### 1. List All Audit Cycles
**GET** `/audit-cycles`

**Permission Required:** `audit_cycles:list`

**Response:**
```json
[
  {
    "id": "uuid",
    "name": "Q1 2024 Audit Cycle",
    "description": "First quarter compliance audits",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-03-31T00:00:00Z",
    "status": "active",
    "created_by": "uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

### 2. Get Audit Cycle by ID
**GET** `/audit-cycles/:id`

**Permission Required:** `audit_cycles:read`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Response:**
```json
{
  "id": "uuid",
  "name": "Q1 2024 Audit Cycle",
  "description": "First quarter compliance audits",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-03-31T00:00:00Z",
  "status": "active",
  "created_by": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### 3. Create Audit Cycle
**POST** `/audit-cycles`

**Permission Required:** `audit_cycles:create`

**Request Body:**
```json
{
  "name": "Q1 2024 Audit Cycle",
  "description": "First quarter compliance audits",
  "start_date": "2024-01-01",
  "end_date": "2024-03-31"
}
```

**Validation:**
- `name` - Required, max 255 characters
- `description` - Optional
- `start_date` - Required, format: YYYY-MM-DD
- `end_date` - Required, format: YYYY-MM-DD, must be >= start_date

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "name": "Q1 2024 Audit Cycle",
  "description": "First quarter compliance audits",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-03-31T00:00:00Z",
  "status": "active",
  "created_by": "uuid",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### 4. Update Audit Cycle
**PUT** `/audit-cycles/:id`

**Permission Required:** `audit_cycles:update`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Request Body:**
```json
{
  "name": "Q1 2024 Audit Cycle - Updated",
  "description": "Updated description",
  "start_date": "2024-01-01",
  "end_date": "2024-03-31",
  "status": "completed"
}
```

**Status Values:**
- `active`
- `completed`
- `archived`

**Response:** `200 OK`

---

### 5. Delete Audit Cycle
**DELETE** `/audit-cycles/:id`

**Permission Required:** `audit_cycles:delete`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Response:** `204 No Content`

**Note:** Cascades to delete all associated clients and framework assignments.

---

### 6. Get Audit Cycle Statistics
**GET** `/audit-cycles/:id/stats`

**Permission Required:** `audit_cycles:read`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Response:**
```json
{
  "id": "uuid",
  "name": "Q1 2024 Audit Cycle",
  "status": "active",
  "total_clients": 5,
  "total_frameworks": 12,
  "completed_frameworks": 3,
  "in_progress_frameworks": 6,
  "pending_frameworks": 2,
  "overdue_frameworks": 1
}
```

---

### 7. Get Clients in Audit Cycle
**GET** `/audit-cycles/:id/clients`

**Permission Required:** `audit_cycles:read`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Response:**
```json
[
  {
    "id": "uuid",
    "audit_cycle_id": "uuid",
    "client_id": "uuid",
    "client_name": "Acme Corp",
    "poc_email": "john@acme.com",
    "client_status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

---

### 8. Add Client to Audit Cycle
**POST** `/audit-cycles/:id/clients`

**Permission Required:** `audit_cycles:manage_clients`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Request Body:**
```json
{
  "client_id": "uuid"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "audit_cycle_id": "uuid",
  "client_id": "uuid",
  "client_name": "Acme Corp",
  "poc_email": "john@acme.com",
  "client_status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `400` - Invalid client_id
- `409` - Client already in audit cycle (unique constraint violation)

---

### 9. Remove Client from Audit Cycle
**DELETE** `/audit-cycles/:id/clients/:clientId`

**Permission Required:** `audit_cycles:manage_clients`

**Parameters:**
- `id` (path) - Audit Cycle UUID
- `clientId` (path) - Client UUID

**Response:** `204 No Content`

**Note:** Cascades to delete all framework assignments for this client in the cycle.

---

### 10. Get Frameworks in Audit Cycle
**GET** `/audit-cycles/:id/frameworks`

**Permission Required:** `audit_cycles:read`

**Parameters:**
- `id` (path) - Audit Cycle UUID

**Response:**
```json
[
  {
    "id": "uuid",
    "audit_cycle_client_id": "uuid",
    "framework_id": "uuid",
    "framework_name": "SOC 2 Type II",
    "client_id": "uuid",
    "client_name": "Acme Corp",
    "assigned_by": "uuid",
    "assigned_at": "2024-01-01T00:00:00Z",
    "due_date": "2024-02-15T00:00:00Z",
    "status": "in_progress",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

---

### 11. Assign Framework to Client in Audit Cycle
**POST** `/audit-cycles/clients/:cycleClientId/frameworks`

**Permission Required:** `audit_cycles:assign_frameworks`

**Parameters:**
- `cycleClientId` (path) - Audit Cycle Client UUID (from step 8)

**Request Body:**
```json
{
  "framework_id": "uuid",
  "framework_name": "SOC 2 Type II",
  "due_date": "2024-02-15"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "audit_cycle_client_id": "uuid",
  "framework_id": "uuid",
  "framework_name": "SOC 2 Type II",
  "client_id": "uuid",
  "client_name": "Acme Corp",
  "assigned_by": "uuid",
  "assigned_at": "2024-01-01T00:00:00Z",
  "due_date": "2024-02-15T00:00:00Z",
  "status": "pending",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## Testing the Endpoints

### Using cURL

#### 1. Create Audit Cycle
```bash
curl -X POST http://localhost:8080/api/audit-cycles \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=YOUR_TOKEN" \
  -d '{
    "name": "Q1 2024 Audit Cycle",
    "description": "First quarter compliance audits",
    "start_date": "2024-01-01",
    "end_date": "2024-03-31"
  }'
```

#### 2. Add Client to Cycle
```bash
curl -X POST http://localhost:8080/api/audit-cycles/CYCLE_ID/clients \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=YOUR_TOKEN" \
  -d '{
    "client_id": "CLIENT_UUID"
  }'
```

#### 3. Assign Framework
```bash
curl -X POST http://localhost:8080/api/audit-cycles/clients/CYCLE_CLIENT_ID/frameworks \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=YOUR_TOKEN" \
  -d '{
    "framework_id": "FRAMEWORK_UUID",
    "framework_name": "SOC 2 Type II",
    "due_date": "2024-02-15"
  }'
```

---

## Frontend Routes

### Pages Created
1. **List Page:** `/audit-cycles`
   - View all audit cycles
   - Search and filter
   - Create new cycle button

2. **Create Page:** `/audit-cycles/new`
   - Form to create new audit cycle
   - Name, description, start/end dates

3. **Detail Page:** `/audit-cycles/:id`
   - View cycle details
   - Statistics cards
   - List of clients in cycle
   - Add client button

4. **Add Client Page:** `/audit-cycles/:id/add-client`
   - Select multiple clients to add
   - Checkbox selection
   - Search functionality
   - Bulk add operation

---

## Error Responses

All endpoints may return these error codes:

- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Duplicate entry (e.g., client already in cycle)
- `500 Internal Server Error` - Server error

**Error Response Format:**
```json
{
  "error": "Error message description"
}
```

---

## Workflow Example

### Complete Audit Cycle Setup

1. **Create Audit Cycle**
   ```
   POST /audit-cycles
   → Returns cycle_id
   ```

2. **Add Clients**
   ```
   POST /audit-cycles/{cycle_id}/clients
   → Returns cycle_client_id for each client
   ```

3. **Assign Frameworks**
   ```
   POST /audit-cycles/clients/{cycle_client_id}/frameworks
   → Creates framework assignment
   ```

4. **Monitor Progress**
   ```
   GET /audit-cycles/{cycle_id}/stats
   → View completion statistics
   ```

5. **Complete Cycle**
   ```
   PUT /audit-cycles/{cycle_id}
   { "status": "completed" }
   ```
