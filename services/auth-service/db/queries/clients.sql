-- name: GetClientByEmailDomain :one
SELECT id, name, email_domain
FROM clients
WHERE email_domain = $1
LIMIT 1;
