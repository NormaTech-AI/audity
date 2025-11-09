-- name: CountClientDatabases :one
SELECT COUNT(*) FROM client_databases;

-- name: CountClientBuckets :one
SELECT COUNT(*) FROM client_buckets;
