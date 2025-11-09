-- name: CreateSubmission :one
INSERT INTO submissions (
    question_id,
    submitted_by,
    answer_value,
    answer_text,
    explanation,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetSubmissionByID :one
SELECT * FROM submissions
WHERE id = $1;

-- name: GetSubmissionByQuestionID :one
SELECT * FROM submissions
WHERE question_id = $1
ORDER BY version DESC
LIMIT 1;

-- name: ListSubmissionsByUser :many
SELECT s.*, q.question_text, q.section
FROM submissions s
JOIN questions q ON q.id = s.question_id
WHERE s.submitted_by = $1
ORDER BY s.submitted_at DESC;

-- name: ListSubmissionsByStatus :many
SELECT s.*, q.question_text, q.section, q.audit_id
FROM submissions s
JOIN questions q ON q.id = s.question_id
WHERE s.status = $1
ORDER BY s.submitted_at DESC;

-- name: UpdateSubmissionAnswer :one
UPDATE submissions
SET 
    answer_value = $2,
    answer_text = $3,
    explanation = $4,
    status = 'in_progress'
WHERE id = $1
RETURNING *;

-- name: SubmitSubmission :one
UPDATE submissions
SET 
    status = 'submitted',
    submitted_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ApproveSubmission :one
UPDATE submissions
SET 
    status = 'approved',
    reviewed_by = $2,
    reviewed_at = NOW(),
    review_notes = $3
WHERE id = $1
RETURNING *;

-- name: RejectSubmission :one
UPDATE submissions
SET 
    status = 'rejected',
    reviewed_by = $2,
    reviewed_at = NOW(),
    rejection_reason = $3,
    review_notes = $4
WHERE id = $1
RETURNING *;

-- name: ReferSubmission :one
UPDATE submissions
SET 
    status = 'referred',
    reviewed_by = $2,
    reviewed_at = NOW(),
    review_notes = $3
WHERE id = $1
RETURNING *;

-- name: ResubmitSubmission :one
INSERT INTO submissions (
    question_id,
    submitted_by,
    answer_value,
    answer_text,
    explanation,
    status,
    version
) VALUES (
    $1, $2, $3, $4, $5, 'submitted',
    (SELECT COALESCE(MAX(version), 0) + 1 FROM submissions WHERE question_id = $1)
)
RETURNING *;

-- name: GetSubmissionWithEvidence :one
SELECT 
    s.*,
    q.question_text,
    q.section,
    COUNT(e.id) as evidence_count
FROM submissions s
JOIN questions q ON q.id = s.question_id
LEFT JOIN evidence e ON e.submission_id = s.id AND e.is_deleted = false
WHERE s.id = $1
GROUP BY s.id, q.id;

-- name: ListPendingReviews :many
SELECT 
    s.*,
    q.question_text,
    q.section,
    q.audit_id,
    a.framework_name
FROM submissions s
JOIN questions q ON q.id = s.question_id
JOIN audits a ON a.id = q.audit_id
WHERE s.status = 'submitted'
ORDER BY s.submitted_at ASC;
