-- name: CreateActivityLog :one
INSERT INTO activity_log (
    user_id,
    user_email,
    action,
    entity_type,
    entity_id,
    details,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: ListActivityLogs :many
SELECT * FROM activity_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActivityLogsByUser :many
SELECT * FROM activity_log
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActivityLogsByEntity :many
SELECT * FROM activity_log
WHERE entity_type = $1 AND entity_id = $2
ORDER BY created_at DESC;

-- name: ListActivityLogsByAction :many
SELECT * FROM activity_log
WHERE action = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentActivity :many
SELECT * FROM activity_log
ORDER BY created_at DESC
LIMIT $1;

-- name: DeleteOldActivityLogs :exec
DELETE FROM activity_log
WHERE created_at < $1;
