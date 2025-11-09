-- name: CreateAudit :one
INSERT INTO audits (
    framework_id,
    framework_name,
    assigned_by,
    assigned_to,
    due_date,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetAuditByID :one
SELECT * FROM audits
WHERE id = $1;

-- name: ListAudits :many
SELECT * FROM audits
ORDER BY created_at DESC;

-- name: ListAuditsByStatus :many
SELECT * FROM audits
WHERE status = $1
ORDER BY due_date ASC;

-- name: UpdateAuditStatus :one
UPDATE audits
SET status = $1,
    completed_at = CASE WHEN $1 = 'completed' THEN NOW() ELSE completed_at END
WHERE id = $2
RETURNING *;

-- name: UpdateAuditAssignee :one
UPDATE audits
SET assigned_to = $1
WHERE id = $2
RETURNING *;

-- name: DeleteAudit :exec
DELETE FROM audits
WHERE id = $1;

-- name: GetAuditProgress :one
SELECT 
    a.id,
    a.framework_name,
    a.status,
    a.due_date,
    COUNT(q.id) as total_questions,
    COUNT(CASE WHEN s.status = 'approved' THEN 1 END) as approved_count,
    COUNT(CASE WHEN s.status = 'submitted' THEN 1 END) as submitted_count,
    COUNT(CASE WHEN s.status = 'rejected' THEN 1 END) as rejected_count
FROM audits a
LEFT JOIN questions q ON q.audit_id = a.id
LEFT JOIN submissions s ON s.question_id = q.id
WHERE a.id = $1
GROUP BY a.id;
