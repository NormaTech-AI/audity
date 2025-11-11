# Audit Cycle Data Model

## Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         AUDIT CYCLES                                │
│─────────────────────────────────────────────────────────────────────│
│ id                UUID (PK)                                          │
│ name              VARCHAR(255)                                       │
│ description       TEXT                                               │
│ start_date        DATE                                               │
│ end_date          DATE                                               │
│ status            VARCHAR(50) [active, completed, archived]          │
│ created_by        UUID (FK -> users.id)                              │
│ created_at        TIMESTAMP                                          │
│ updated_at        TIMESTAMP                                          │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ 1
                              │
                              │ has many
                              │
                              │ N
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    AUDIT CYCLE CLIENTS                              │
│─────────────────────────────────────────────────────────────────────│
│ id                UUID (PK)                                          │
│ audit_cycle_id    UUID (FK -> audit_cycles.id)                      │
│ client_id         UUID (FK -> clients.id)                           │
│ created_at        TIMESTAMP                                          │
│                                                                      │
│ UNIQUE(audit_cycle_id, client_id)                                   │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ 1
                              │
                              │ has many
                              │
                              │ N
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  AUDIT CYCLE FRAMEWORKS                             │
│─────────────────────────────────────────────────────────────────────│
│ id                     UUID (PK)                                     │
│ audit_cycle_client_id  UUID (FK -> audit_cycle_clients.id)          │
│ framework_id           UUID (References framework-service)           │
│ framework_name         VARCHAR(255)                                  │
│ assigned_by            UUID (FK -> users.id)                         │
│ assigned_at            TIMESTAMP                                     │
│ due_date               DATE                                          │
│ status                 VARCHAR(50) [pending, in_progress,            │
│                                     completed, overdue]              │
│ created_at             TIMESTAMP                                     │
│ updated_at             TIMESTAMP                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## Relationships

### 1. Audit Cycle → Audit Cycle Clients (One-to-Many)
- One audit cycle can have many clients
- Cascade delete: Deleting a cycle removes all its client associations

### 2. Audit Cycle Clients → Audit Cycle Frameworks (One-to-Many)
- One client in a cycle can have many frameworks assigned
- Cascade delete: Removing a client from a cycle removes all its framework assignments

### 3. External References
- `audit_cycles.created_by` → `users.id` (SET NULL on delete)
- `audit_cycle_clients.client_id` → `clients.id` (CASCADE on delete)
- `audit_cycle_frameworks.assigned_by` → `users.id` (SET NULL on delete)
- `audit_cycle_frameworks.framework_id` → framework-service (external reference)

## Workflow

```
1. Create Audit Cycle
   ↓
2. Add Clients to Cycle
   ↓
3. Assign Frameworks to Each Client
   ↓
4. Track Framework Completion Status
   ↓
5. Monitor Progress via Statistics
   ↓
6. Complete or Archive Cycle
```

## Example Data Flow

```
Audit Cycle: "Q1 2024 Compliance Audit"
├── Client: "Acme Corp"
│   ├── Framework: "SOC 2 Type II" (status: in_progress)
│   ├── Framework: "ISO 27001" (status: pending)
│   └── Framework: "GDPR" (status: completed)
│
├── Client: "TechStart Inc"
│   ├── Framework: "SOC 2 Type II" (status: pending)
│   └── Framework: "HIPAA" (status: in_progress)
│
└── Client: "Global Finance Ltd"
    ├── Framework: "PCI DSS" (status: completed)
    └── Framework: "SOC 2 Type II" (status: overdue)
```

## Statistics Aggregation

The `GetAuditCycleStats` query provides:
- **total_clients**: Count of unique clients in the cycle
- **total_frameworks**: Total framework assignments across all clients
- **completed_frameworks**: Count of frameworks with status = 'completed'
- **in_progress_frameworks**: Count of frameworks with status = 'in_progress'
- **pending_frameworks**: Count of frameworks with status = 'pending'
- **overdue_frameworks**: Count of frameworks with status = 'overdue'

## Indexes

### audit_cycles
- `idx_audit_cycles_status` on (status)
- `idx_audit_cycles_dates` on (start_date, end_date)

### audit_cycle_clients
- `idx_audit_cycle_clients_cycle_id` on (audit_cycle_id)
- `idx_audit_cycle_clients_client_id` on (client_id)

### audit_cycle_frameworks
- `idx_audit_cycle_frameworks_cycle_client_id` on (audit_cycle_client_id)
- `idx_audit_cycle_frameworks_status` on (status)

## Constraints

1. **Date Validation**: `end_date >= start_date` in audit_cycles
2. **Status Validation**: 
   - audit_cycles.status IN ('active', 'completed', 'archived')
   - audit_cycle_frameworks.status IN ('pending', 'in_progress', 'completed', 'overdue')
3. **Unique Client Assignment**: UNIQUE(audit_cycle_id, client_id) prevents duplicate client assignments
4. **Referential Integrity**: Foreign keys with appropriate CASCADE/SET NULL actions
