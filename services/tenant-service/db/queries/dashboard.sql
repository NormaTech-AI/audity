-- name: CountTotalUsers :one
SELECT COUNT(*) FROM users;

-- name: CountClientFrameworks :one
SELECT COUNT(*) FROM client_frameworks;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs;
