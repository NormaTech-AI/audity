-- name: CreateEvidence :one
INSERT INTO evidence (
    submission_id,
    file_name,
    file_path,
    file_size,
    file_type,
    uploaded_by,
    description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetEvidenceByID :one
SELECT * FROM evidence
WHERE id = $1 AND is_deleted = false;

-- name: ListEvidenceBySubmission :many
SELECT * FROM evidence
WHERE submission_id = $1 AND is_deleted = false
ORDER BY uploaded_at DESC;

-- name: ListEvidenceByUser :many
SELECT e.*, s.question_id
FROM evidence e
JOIN submissions s ON s.id = e.submission_id
WHERE e.uploaded_by = $1 AND e.is_deleted = false
ORDER BY e.uploaded_at DESC;

-- name: SoftDeleteEvidence :one
UPDATE evidence
SET 
    is_deleted = true,
    deleted_at = NOW(),
    deleted_by = $2
WHERE id = $1
RETURNING *;

-- name: HardDeleteEvidence :exec
DELETE FROM evidence
WHERE id = $1;

-- name: GetEvidenceStats :one
SELECT 
    COUNT(*) as total_files,
    SUM(file_size) as total_size,
    COUNT(DISTINCT submission_id) as submissions_with_evidence
FROM evidence
WHERE is_deleted = false;
