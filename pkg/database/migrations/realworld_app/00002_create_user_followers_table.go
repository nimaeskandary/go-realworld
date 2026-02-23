package realworld_app

import (
	"context"
	"database/sql"

	domain "github.com/nimaeskandary/go-realworld/pkg/database/types"
)

// createFolloweTableMigration - migration for creating followers table.
// This is done as a go migration as an example.
// Sql based migrations are better for most use cases that don't require complex logic.
// An example use case for a go migration is, say you stored data as a protobuf blob under a data column,
// and that proto schema changed. You can write a go migration that reads the old proto blob and re
// encodes it using the new proto schema version.
type createUserFollowersTableMigration struct{}

func init() {
	registerCodeMigration(&createUserFollowersTableMigration{})
}

func (m *createUserFollowersTableMigration) Version() int64 {
	return 2
}

func (m *createUserFollowersTableMigration) Up() domain.MigrationFn {
	return func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS user_followers (
				followed_by_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				following_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

				-- index to optimize for fetching users that a user is following
				PRIMARY KEY (followed_by_user_id, following_user_id),
				
				-- prevent self follows
				CONSTRAINT followers_no_self_follow CHECK (followed_by_user_id != following_user_id)
			);
		`)
		return err
	}
}

func (m *createUserFollowersTableMigration) Down() domain.MigrationFn {
	return func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			DROP TABLE IF EXISTS user_followers;
		`)
		return err
	}
}
