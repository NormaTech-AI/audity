-- name: GetUserByOIDC :one
SELECT id, email, name, designation, client_id, created_at, updated_at, last_login
FROM users
WHERE oidc_provider = $1 AND oidc_sub = $2
LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, name, designation, client_id, created_at, updated_at, last_login
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, name, oidc_provider, oidc_sub, designation, client_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, name, designation, client_id, created_at, updated_at, last_login;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = NOW()
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, name, designation, client_id, created_at, updated_at, last_login
FROM users
WHERE email = $1
LIMIT 1;

-- name: ListUsers :many
SELECT id, email, name, designation, client_id, created_at, updated_at, last_login
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = $3, designation = $4
WHERE id = $1
RETURNING id, email, name, designation, client_id, created_at, updated_at, last_login;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUserClientID :exec
UPDATE users
SET client_id = $2
WHERE id = $1;
