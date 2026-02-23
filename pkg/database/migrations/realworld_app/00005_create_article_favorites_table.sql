-- +goose Up
CREATE TABLE IF NOT EXISTS article_favorites (
    article_id TEXT REFERENCES articles(id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (article_id, user_id)
);
CREATE INDEX idx_article_favorites_user_id ON article_favorites(user_id);

-- +goose Down
DROP TABLE IF EXISTS article_favorites;
