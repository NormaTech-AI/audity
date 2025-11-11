-- name: CreateAuditCycle :one
INSERT INTO audit_cycles (name, description, start_date, end_date, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAuditCycle :one
SELECT * FROM audit_cycles
WHERE id = $1 LIMIT 1;

-- name: ListAuditCycles :many
SELECT * FROM audit_cycles
ORDER BY created_at DESC;

-- name: ListAuditCyclesByStatus :many
SELECT * FROM audit_cycles
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateAuditCycle :one
UPDATE audit_cycles
SET name = COALESCE($2, name),
    description = COALESCE($3, description),
    start_date = COALESCE($4, start_date),
    end_date = COALESCE($5, end_date),
    status = COALESCE($6, status)
WHERE id = $1
RETURNING *;

-- name: DeleteAuditCycle :exec
DELETE FROM audit_cycles
WHERE id = $1;

-- name: AddClientToAuditCycle :one
INSERT INTO audit_cycle_clients (audit_cycle_id, client_id)
VALUES ($1, $2)
RETURNING *;

-- name: RemoveClientFromAuditCycle :exec
DELETE FROM audit_cycle_clients
WHERE audit_cycle_id = $1 AND client_id = $2;

-- name: GetAuditCycleClients :many
SELECT 
    acc.id,
    acc.audit_cycle_id,
    acc.client_id,
    acc.created_at,
    c.name as client_name,
    c.poc_email,
    c.status as client_status
FROM audit_cycle_clients acc
JOIN clients c ON acc.client_id = c.id
WHERE acc.audit_cycle_id = $1
ORDER BY c.name ASC;

-- name: AssignFrameworkToAuditCycleClient :one
INSERT INTO audit_cycle_frameworks (
    audit_cycle_client_id,
    framework_id,
    framework_name,
    assigned_by,
    due_date,
    status,
    auditor_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAuditCycleFrameworks :many
SELECT 
    acf.id,
    acf.audit_cycle_client_id,
    acf.framework_id,
    acf.framework_name,
    acf.assigned_by,
    acf.assigned_at,
    acf.due_date,
    acf.status,
    acf.auditor_id,
    acf.created_at,
    acf.updated_at,
    acc.client_id,
    c.name as client_name
FROM audit_cycle_frameworks acf
JOIN audit_cycle_clients acc ON acf.audit_cycle_client_id = acc.id
JOIN clients c ON acc.client_id = c.id
WHERE acc.audit_cycle_id = $1
ORDER BY c.name ASC, acf.framework_name ASC;

-- name: GetClientFrameworksInCycle :many
SELECT 
    acf.id,
    acf.audit_cycle_client_id,
    acf.framework_id,
    acf.framework_name,
    acf.assigned_by,
    acf.assigned_at,
    acf.due_date,
    acf.status,
    acf.auditor_id,
    acf.created_at,
    acf.updated_at
FROM audit_cycle_frameworks acf
WHERE acf.audit_cycle_client_id = $1
ORDER BY acf.framework_name ASC;

-- name: UpdateAuditCycleFrameworkStatus :one
UPDATE audit_cycle_frameworks
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAuditCycleFramework :exec
DELETE FROM audit_cycle_frameworks
WHERE id = $1;

-- name: GetAuditCycleStats :one
SELECT 
    ac.id,
    ac.name,
    ac.status,
    COUNT(DISTINCT acc.client_id) as total_clients,
    COUNT(acf.id) as total_frameworks,
    COUNT(CASE WHEN acf.status = 'completed' THEN 1 END) as completed_frameworks,
    COUNT(CASE WHEN acf.status = 'in_progress' THEN 1 END) as in_progress_frameworks,
    COUNT(CASE WHEN acf.status = 'pending' THEN 1 END) as pending_frameworks,
    COUNT(CASE WHEN acf.status = 'overdue' THEN 1 END) as overdue_frameworks
FROM audit_cycles ac
LEFT JOIN audit_cycle_clients acc ON ac.id = acc.audit_cycle_id
LEFT JOIN audit_cycle_frameworks acf ON acc.id = acf.audit_cycle_client_id
WHERE ac.id = $1
GROUP BY ac.id, ac.name, ac.status;
