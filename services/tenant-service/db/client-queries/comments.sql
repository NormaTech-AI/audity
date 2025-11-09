-- name: CreateComment :one
INSERT INTO comments (
    submission_id,
    user_id,
    user_name,
    comment_text,
    is_internal
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1;

-- name: ListCommentsBySubmission :many
SELECT * FROM comments
WHERE submission_id = $1
ORDER BY created_at ASC;

-- name: ListInternalComments :many
SELECT * FROM comments
WHERE submission_id = $1 AND is_internal = true
ORDER BY created_at ASC;

-- name: ListExternalComments :many
SELECT * FROM comments
WHERE submission_id = $1 AND is_internal = false
ORDER BY created_at ASC;

-- name: UpdateComment :one
UPDATE comments
SET comment_text = $2
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;
