package db_types

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/nimaeskandary/go-realworld/pkg/util"
)

type SqlMigrationRunner interface {
	util.FxLifecycle
	Apply(ctx context.Context, version int64) error
	ApplyAll(ctx context.Context) error
	Rollback(ctx context.Context, version int64) error
	RollbackTo(ctx context.Context, version int64) error
	Status(ctx context.Context) (string, error)
	CurrentVersion(ctx context.Context) (int64, error)
}

type MigrationFn func(ctx context.Context, tx *sql.Tx) error

// Supporting code based migrations
type GoMigration interface {
	Version() int64
	Up() MigrationFn
	Down() MigrationFn
}

type MigrationsProvider interface {
	GetCodeMigrations() *[]GoMigration
	GetSqlMigrationsFs() fs.FS
}
