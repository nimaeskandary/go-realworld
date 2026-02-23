-- +goose Up
CREATE TABLE IF NOT EXISTS article_tags (
    article_id TEXT REFERENCES articles(id) ON DELETE CASCADE,
    tag TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (article_id, tag)
);
CREATE INDEX idx_article_tags_tag ON article_tags(tag);

-- +goose Down
DROP TABLE IF EXISTS article_tags;
