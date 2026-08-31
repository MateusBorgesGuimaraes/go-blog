-- name: ListPublishedPosts :many
SELECT id, title, slug, excerpt, cover_image_url, status, author_id, published_at, created_at, updated_at
FROM posts
WHERE status = 'published'
ORDER BY published_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPostBySlug :one
SELECT
    p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url,
    p.status, p.author_id, p.published_at, p.created_at, p.updated_at,
    u.name AS author_name
FROM posts p
INNER JOIN users u ON u.id = p.author_id
WHERE p.slug = $1;

-- name: ListAllPosts :many
SELECT id, title, slug, excerpt, cover_image_url, status, author_id, published_at, created_at, updated_at
FROM posts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetPostByID :one
SELECT
    p.id, p.title, p.slug, p.content, p.excerpt, p.cover_image_url,
    p.status, p.author_id, p.published_at, p.created_at, p.updated_at,
    u.name AS author_name
FROM posts p
INNER JOIN users u ON u.id = p.author_id
WHERE p.id = $1;

-- name: CreatePost :one
INSERT INTO posts (title, slug, content, excerpt, cover_image_url, status, author_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, title, slug, content, excerpt, cover_image_url, status, author_id, published_at, created_at, updated_at;

-- name: UpdatePost :one
UPDATE posts
SET title = $2,
    content = $3,
    excerpt = $4,
    cover_image_url = $5,
    updated_at = now()
WHERE id = $1
RETURNING id, title, slug, content, excerpt, cover_image_url, status, author_id, published_at, created_at, updated_at;

-- name: PublishPost :one
UPDATE posts
SET status = 'published',
    published_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING id, title, slug, content, excerpt, cover_image_url, status, author_id, published_at, created_at, updated_at;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;
