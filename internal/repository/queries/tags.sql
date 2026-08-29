-- name: ListTags :many
SELECT id, name, slug
FROM tags
ORDER BY name ASC;

-- name: GetTagBySlug :one
SELECT id, name, slug
FROM tags
WHERE slug = $1;

-- name: CreateTag :one
INSERT INTO tags (name, slug)
VALUES ($1, $2)
RETURNING id, name, slug;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1;

-- name: AddTagToPost :exec
INSERT INTO post_tags (post_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromPost :exec
DELETE FROM post_tags
WHERE post_id = $1 AND tag_id = $2;

-- name: ListTagsByPostID :many
SELECT t.id, t.name, t.slug
FROM tags t
INNER JOIN post_tags pt ON pt.tag_id = t.id
WHERE pt.post_id = $1;
