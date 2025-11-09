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
