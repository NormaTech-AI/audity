-- name: CreateClientDatabase :one
INSERT INTO client_databases (client_id, db_name, db_host, db_port, db_user, encrypted_password)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetClientDatabase :one
SELECT * FROM client_databases
WHERE client_id = $1 LIMIT 1;

-- name: GetClientDatabaseByName :one
SELECT * FROM client_databases
WHERE db_name = $1 LIMIT 1;

-- name: ListClientDatabases :many
SELECT * FROM client_databases
ORDER BY created_at DESC;

-- name: DeleteClientDatabase :exec
DELETE FROM client_databases
WHERE client_id = $1;
