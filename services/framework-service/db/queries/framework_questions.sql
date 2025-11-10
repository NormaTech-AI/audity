-- name: CreateFrameworkQuestion :one
INSERT INTO framework_questions (
    framework_id,
    section_title,
    control_id,
    question_text,
    help_text,
    acceptable_evidence
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: BulkCreateFrameworkQuestions :copyfrom
INSERT INTO framework_questions (
    framework_id,
    control_id,
    question_text,
    help_text,
    acceptable_evidence
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetFrameworkQuestion :one
SELECT * FROM framework_questions
WHERE question_id = $1 LIMIT 1;

-- name: ListFrameworkQuestions :many
SELECT * FROM framework_questions
WHERE framework_id = $1
ORDER BY control_id;

-- name: UpdateFrameworkQuestion :one
UPDATE framework_questions
SET 
    section_title = $2,
    control_id = $3,
    question_text = $4,
    help_text = $5,
    acceptable_evidence = $6
WHERE question_id = $1
RETURNING *;

-- name: DeleteFrameworkQuestion :exec
DELETE FROM framework_questions
WHERE question_id = $1;

-- name: DeleteFrameworkQuestionsByFrameworkId :exec
DELETE FROM framework_questions
WHERE framework_id = $1;

-- name: CountFrameworkQuestions :one
SELECT COUNT(*) FROM framework_questions
WHERE framework_id = $1;

-- name: GetFrameworkWithQuestions :many
SELECT 
    f.id as framework_id,
    f.name as framework_name,
    f.description as framework_description,
    f.version as framework_version,
    fq.question_id,
    fq.control_id,
    fq.question_text,
    fq.help_text,
    fq.acceptable_evidence
FROM compliance_frameworks f
LEFT JOIN framework_questions fq ON f.id = fq.framework_id
WHERE f.id = $1
ORDER BY fq.control_id;
