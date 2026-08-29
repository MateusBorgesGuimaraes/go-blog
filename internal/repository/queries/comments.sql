-- name: ListApprovedCommentsByPostID :many
SELECT id, post_id, author_name, content, status, created_at
FROM comments
WHERE post_id = $1 AND status = 'approved'
ORDER BY created_at ASC;

-- name: ListAllCommentsByPostID :many
SELECT id, post_id, author_name, content, status, created_at
FROM comments
WHERE post_id = $1
ORDER BY created_at DESC;

-- name: CreateComment :one
INSERT INTO comments (post_id, author_name, content, status)
VALUES ($1, $2, $3, 'pending')
RETURNING id, post_id, author_name, content, status, created_at;

-- name: ApproveComment :one
UPDATE comments
SET status = 'approved'
WHERE id = $1
RETURNING id, post_id, author_name, content, status, created_at;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;
