CREATE TABLE posts (
    id               SERIAL PRIMARY KEY,
    title            VARCHAR(255) NOT NULL,
    slug             VARCHAR(255) NOT NULL UNIQUE,
    content          TEXT NOT NULL,
    excerpt          VARCHAR(500),
    cover_image_url  VARCHAR(500),
    status           VARCHAR(20) NOT NULL DEFAULT 'draft',
    author_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_author_id ON posts(author_id);
