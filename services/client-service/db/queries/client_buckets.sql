-- name: CreateClientBucket :one
INSERT INTO client_buckets (client_id, bucket_name)
VALUES ($1, $2)
RETURNING *;

-- name: GetClientBucket :one
SELECT * FROM client_buckets
WHERE client_id = $1 LIMIT 1;

-- name: GetClientBucketByName :one
SELECT * FROM client_buckets
WHERE bucket_name = $1 LIMIT 1;

-- name: ListClientBuckets :many
SELECT * FROM client_buckets
ORDER BY created_at DESC;

-- name: DeleteClientBucket :exec
DELETE FROM client_buckets
WHERE client_id = $1;
