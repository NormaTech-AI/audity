-- name: CreateQuestion :one
INSERT INTO questions (
    audit_id,
    section,
    question_number,
    question_text,
    question_type,
    help_text,
    is_mandatory,
    display_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions
WHERE id = $1;

-- name: ListQuestionsByAudit :many
SELECT * FROM questions
WHERE audit_id = $1
ORDER BY display_order ASC;

-- name: ListQuestionsBySection :many
SELECT * FROM questions
WHERE audit_id = $1 AND section = $2
ORDER BY display_order ASC;

-- name: UpdateQuestion :one
UPDATE questions
SET 
    question_text = COALESCE($2, question_text),
    help_text = COALESCE($3, help_text),
    is_mandatory = COALESCE($4, is_mandatory)
WHERE id = $1
RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions
WHERE id = $1;

-- name: BulkCreateQuestions :copyfrom
INSERT INTO questions (
    audit_id,
    section,
    question_number,
    question_text,
    question_type,
    help_text,
    is_mandatory,
    display_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetQuestionWithSubmission :one
SELECT 
    q.*,
    s.id as submission_id,
    s.answer_value,
    s.answer_text,
    s.explanation,
    s.status as submission_status,
    s.submitted_at,
    s.reviewed_by,
    s.reviewed_at,
    s.review_notes
FROM questions q
LEFT JOIN submissions s ON s.question_id = q.id
WHERE q.id = $1;

-- name: ListQuestionsWithSubmissions :many
SELECT 
    q.*,
    s.id as submission_id,
    s.status as submission_status,
    s.submitted_at
FROM questions q
LEFT JOIN submissions s ON s.question_id = q.id
WHERE q.audit_id = $1
ORDER BY q.display_order ASC;

-- name: ListQuestionsForUser :many
-- Get questions for a specific user based on their role
-- POC users see all questions, stakeholders see only assigned questions
SELECT DISTINCT
    q.*,
    s.id as submission_id,
    s.answer_value,
    s.answer_text,
    s.explanation,
    s.status as submission_status,
    s.submitted_at,
    s.submitted_by,
    qa.assigned_to as assigned_user_id
FROM questions q
LEFT JOIN submissions s ON s.question_id = q.id
LEFT JOIN question_assignments qa ON qa.question_id = q.id
WHERE q.audit_id = $1
    AND (
        $2 = true  -- is_poc: if true, show all questions
        OR qa.assigned_to = $3  -- if stakeholder, show only assigned questions
    )
ORDER BY q.display_order ASC;

-- name: AssignQuestionToUser :one
INSERT INTO question_assignments (
    question_id,
    assigned_to,
    assigned_by,
    notes
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (question_id, assigned_to) DO UPDATE
SET notes = EXCLUDED.notes
RETURNING *;

-- name: UnassignQuestionFromUser :exec
DELETE FROM question_assignments
WHERE question_id = $1 AND assigned_to = $2;

-- name: ListQuestionAssignments :many
SELECT * FROM question_assignments
WHERE question_id = $1;

-- name: ListUserAssignments :many
SELECT 
    qa.*,
    q.question_text,
    q.section,
    q.audit_id,
    a.framework_name
FROM question_assignments qa
JOIN questions q ON q.id = qa.question_id
JOIN audits a ON a.id = q.audit_id
WHERE qa.assigned_to = $1
ORDER BY qa.assigned_at DESC;
