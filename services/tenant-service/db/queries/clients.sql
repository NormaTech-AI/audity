-- name: CreateClient :one
INSERT INTO clients (name, poc_email, status, email_domain)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1 LIMIT 1;

-- name: GetClientByEmail :one
SELECT * FROM clients
WHERE poc_email = $1 LIMIT 1;

-- name: GetClientByEmailDomain :one
SELECT * FROM clients
WHERE email_domain = $1 LIMIT 1;

-- name: ListClients :many
SELECT * FROM clients
ORDER BY created_at DESC;

-- name: ListActiveClients :many
SELECT * FROM clients
WHERE status = 'active'
ORDER BY created_at DESC;

-- name: UpdateClient :one
UPDATE clients
SET name = $2, poc_email = $3, status = $4, email_domain = $5
WHERE id = $1
RETURNING *;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1;

-- name: CountClients :one
SELECT COUNT(*) FROM clients;

-- name: CountClientsByStatus :one
SELECT COUNT(*) FROM clients
WHERE status = $1;
