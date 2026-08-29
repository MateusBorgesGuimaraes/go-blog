CREATE TABLE comments (
    id           SERIAL PRIMARY KEY,
    post_id      INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_name  VARCHAR(255) NOT NULL,
    content      TEXT NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | approved | spam
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_status ON comments(status);
