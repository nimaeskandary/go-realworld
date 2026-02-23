package internal_test

import (
	"context"
	"database/sql"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/nimaeskandary/go-realworld/pkg/database"
	"github.com/nimaeskandary/go-realworld/pkg/database/internal"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	db_config_provider "github.com/nimaeskandary/go-realworld/pkg/test_utils/db_config_provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GooseMigrationRunner(t *testing.T) {
	testCfg := db_types.SqlDbConfig{
		Host:     "localhost",
		Port:     15432,
		Username: "testpostgres",
		Password: "testpassword",
		DBName:   "goose_migration_runner_test",
		SslMode:  "disable",
	}

	provider := db_config_provider.NewPostgresSqlDbConfigProvider(testCfg,
		// no need for migrations in the database provider, that is tested elsewhere
		func(_ context.Context, _ db_types.SQLDatabase) error {
			return nil
		})

	t.Cleanup(func() { _ = provider.Cleanup(context.Background(), testCfg) })

	mockFs := fstest.MapFS{
		"001_create_users.sql": {
			Data: []byte(`
-- +goose Up
CREATE TABLE users (id SERIAL PRIMARY KEY, username TEXT);
-- +goose Down
DROP TABLE users;
			`),
		},
	}

	codeMigration := &StubGoMigration{
		version: 2,
		up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "INSERT INTO users (username) VALUES ('testuser')")
			return err
		},
		down: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE username = 'testuser'")
			return err
		},
	}

	mockProvider := &StubMigrationsProvider{
		fs:             mockFs,
		codeMigrations: []db_types.GoMigration{codeMigration},
	}

	t.Run("Full Migration Lifecycle", func(t *testing.T) {
		dbCfg, err := provider.GetFreshDbConfig(t.Context())
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = provider.Cleanup(context.Background(), dbCfg)
		})

		// setup db connection
		db, err := database.NewPostgresSQLDatabase(dbCfg)
		require.NoError(t, err)
		require.NoError(t, db.Start(t.Context()))
		defer func() { _ = db.Stop(t.Context()) }()

		// setup migration runner
		runner, err := internal.NewGooseMigrationRunner(db, mockProvider)
		require.NoError(t, err)
		require.NoError(t, runner.Start(t.Context()))
		defer func() { _ = runner.Stop(t.Context()) }()

		// apply all migrations
		err = runner.ApplyAll(t.Context())
		require.NoError(t, err, "Failed to apply migrations")

		// verify migrations ran
		var username string
		err = db.GetDB().QueryRowContext(t.Context(), "SELECT username FROM users WHERE username = 'testuser'").Scan(&username)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", username)

		// verify version
		ver, err := runner.CurrentVersion(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), ver)

		// rollback last migration
		err = runner.Rollback(t.Context(), 2)
		require.NoError(t, err)

		// verify data is gone
		err = db.GetDB().QueryRowContext(t.Context(), "SELECT username FROM users WHERE username = 'testuser'").Scan(&username)
		assert.Equal(t, err, sql.ErrNoRows)

		// verify table still exists
		tableName := new(string)
		err = db.GetDB().QueryRowContext(t.Context(), "SELECT tablename FROM pg_tables WHERE tablename = 'users'").Scan(&tableName)
		assert.NoError(t, err)

		// rollback all migrations
		err = runner.RollbackTo(t.Context(), 0)
		require.NoError(t, err)

		// verify table is gone
		err = db.GetDB().QueryRowContext(t.Context(), "SELECT tablename FROM pg_tables WHERE tablename = 'users'").Scan(&tableName)
		assert.Equal(t, err, sql.ErrNoRows)
	})
}

type StubGoMigration struct {
	version int64
	up      db_types.MigrationFn
	down    db_types.MigrationFn
}

func (m *StubGoMigration) Version() int64             { return m.version }
func (m *StubGoMigration) Up() db_types.MigrationFn   { return m.up }
func (m *StubGoMigration) Down() db_types.MigrationFn { return m.down }

type StubMigrationsProvider struct {
	fs             fs.FS
	codeMigrations []db_types.GoMigration
}

func (p *StubMigrationsProvider) GetSqlMigrationsFs() fs.FS {
	return p.fs
}

func (p *StubMigrationsProvider) GetCodeMigrations() *[]db_types.GoMigration {
	return &p.codeMigrations
}
