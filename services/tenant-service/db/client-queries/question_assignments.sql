-- name: CreateQuestionAssignment :one
INSERT INTO question_assignments (
    question_id,
    assigned_to,
    assigned_by,
    notes
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetQuestionAssignment :one
SELECT * FROM question_assignments
WHERE question_id = $1 AND assigned_to = $2;

-- name: ListAssignmentsByUser :many
SELECT 
    qa.*,
    q.question_text,
    q.section,
    q.audit_id,
    s.status as submission_status
FROM question_assignments qa
JOIN questions q ON q.id = qa.question_id
LEFT JOIN submissions s ON s.question_id = q.id
WHERE qa.assigned_to = $1
ORDER BY qa.assigned_at DESC;

-- name: ListAssignmentsByQuestion :many
SELECT * FROM question_assignments
WHERE question_id = $1
ORDER BY assigned_at DESC;

-- name: DeleteQuestionAssignment :exec
DELETE FROM question_assignments
WHERE question_id = $1 AND assigned_to = $2;

-- name: BulkAssignQuestions :copyfrom
INSERT INTO question_assignments (
    question_id,
    assigned_to,
    assigned_by,
    notes
) VALUES (
    $1, $2, $3, $4
);
