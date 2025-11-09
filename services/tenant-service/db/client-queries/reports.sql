-- name: CreateReport :one
INSERT INTO reports (
    audit_id,
    unsigned_file_path,
    generated_by,
    status
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetReportByID :one
SELECT * FROM reports
WHERE id = $1;

-- name: GetReportByAuditID :one
SELECT * FROM reports
WHERE audit_id = $1;

-- name: UpdateReportUnsigned :one
UPDATE reports
SET 
    unsigned_file_path = $2,
    status = 'generated'
WHERE id = $1
RETURNING *;

-- name: UpdateReportSigned :one
UPDATE reports
SET 
    signed_file_path = $2,
    signed_by = $3,
    signed_at = NOW(),
    status = 'signed'
WHERE id = $1
RETURNING *;

-- name: MarkReportDelivered :one
UPDATE reports
SET status = 'delivered'
WHERE id = $1
RETURNING *;

-- name: ListReportsByStatus :many
SELECT 
    r.*,
    a.framework_name,
    a.status as audit_status
FROM reports r
JOIN audits a ON a.id = r.audit_id
WHERE r.status = $1
ORDER BY r.generated_at DESC;

-- name: DeleteReport :exec
DELETE FROM reports
WHERE id = $1;
