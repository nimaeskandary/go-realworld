-- +goose Up
CREATE TABLE IF NOT EXISTS articles (
    id TEXT PRIMARY KEY,
    author_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS articles;
