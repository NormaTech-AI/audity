-- name: AssignFrameworkToClient :one
INSERT INTO client_frameworks (client_id, framework_id, due_date, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetClientFramework :one
SELECT 
    cf.id,
    cf.client_id,
    cf.framework_id,
    cf.due_date,
    cf.status,
    cf.created_at,
    cf.updated_at,
    c.name as client_name,
    f.name as framework_name
FROM client_frameworks cf
JOIN clients c ON cf.client_id = c.id
JOIN compliance_frameworks f ON cf.framework_id = f.id
WHERE cf.id = $1 LIMIT 1;

-- name: ListClientFrameworks :many
SELECT 
    cf.id,
    cf.client_id,
    cf.framework_id,
    cf.due_date,
    cf.status,
    cf.created_at,
    cf.updated_at,
    c.name as client_name,
    f.name as framework_name
FROM client_frameworks cf
JOIN clients c ON cf.client_id = c.id
JOIN compliance_frameworks f ON cf.framework_id = f.id
WHERE cf.client_id = $1
ORDER BY cf.created_at DESC;

-- name: ListFrameworksByStatus :many
SELECT 
    cf.id,
    cf.client_id,
    cf.framework_id,
    cf.due_date,
    cf.status,
    cf.created_at,
    cf.updated_at,
    c.name as client_name,
    f.name as framework_name
FROM client_frameworks cf
JOIN clients c ON cf.client_id = c.id
JOIN compliance_frameworks f ON cf.framework_id = f.id
WHERE cf.status = $1
ORDER BY cf.due_date ASC;

-- name: UpdateClientFrameworkStatus :one
UPDATE client_frameworks
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteClientFramework :exec
DELETE FROM client_frameworks
WHERE id = $1;
