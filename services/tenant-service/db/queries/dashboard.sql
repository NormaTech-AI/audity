-- name: CountTotalUsers :one
SELECT COUNT(*) FROM users;

-- name: CountClientFrameworks :one
SELECT COUNT(*) FROM client_frameworks;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs;

-- Client-specific dashboard queries

-- name: GetClientAuditCycleEnrollments :many
-- Get all audit cycles a specific client is enrolled in with framework details
SELECT 
    ac.id as audit_cycle_id,
    ac.name as audit_cycle_name,
    ac.description as audit_cycle_description,
    ac.start_date,
    ac.end_date,
    ac.status as cycle_status,
    acc.id as enrollment_id,
    acc.created_at as enrolled_at,
    acf.id as framework_assignment_id,
    acf.framework_id,
    acf.framework_name,
    acf.due_date,
    acf.status as framework_status,
    acf.auditor_id
FROM audit_cycle_clients acc
JOIN audit_cycles ac ON acc.audit_cycle_id = ac.id
LEFT JOIN audit_cycle_frameworks acf ON acc.id = acf.audit_cycle_client_id
WHERE acc.client_id = $1
ORDER BY ac.start_date DESC, acf.framework_name ASC;

-- name: CountClientActiveAuditCycles :one
-- Count active audit cycles for a specific client
SELECT COUNT(DISTINCT ac.id)
FROM audit_cycle_clients acc
JOIN audit_cycles ac ON acc.audit_cycle_id = ac.id
WHERE acc.client_id = $1 AND ac.status = 'active';

-- name: CountClientTotalFrameworkAssignments :one
-- Count total framework assignments across all audit cycles for a client
SELECT COUNT(acf.id)
FROM audit_cycle_clients acc
LEFT JOIN audit_cycle_frameworks acf ON acc.id = acf.audit_cycle_client_id
WHERE acc.client_id = $1;
