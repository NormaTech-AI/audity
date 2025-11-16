-- name: CreateFramework :one
INSERT INTO compliance_frameworks (name, description, version)
VALUES ($1, $2, $3)
RETURNING id, name, description, version, created_at, updated_at;

-- name: GetFramework :one
SELECT * FROM compliance_frameworks
WHERE id = $1 LIMIT 1;

-- name: GetFrameworkByName :one
SELECT * FROM compliance_frameworks
WHERE name = $1 LIMIT 1;

-- name: ListFrameworks :many
SELECT * FROM compliance_frameworks
ORDER BY created_at DESC;

-- name: UpdateFramework :one
UPDATE compliance_frameworks
SET name = $2, description = $3, version = $5
WHERE id = $1
RETURNING *;

-- name: DeleteFramework :exec
DELETE FROM compliance_frameworks
WHERE id = $1;

-- name: CountFrameworks :one
SELECT COUNT(*) FROM compliance_frameworks;
