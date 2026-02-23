package realworld_app

import (
	"embed"
	"io/fs"
	"sync"

	"github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

// bundles .sql files to be used as an embedded FileSystem by the migration runner
//
//go:embed *.sql
var sqlMigrations embed.FS
var codeMigrations = &[]db_types.GoMigration{}
var lock = &sync.Mutex{}

type migrationsProviderImpl struct{}

// create a singleton instance of the migrations provider
var p = &migrationsProviderImpl{}

func NewMigrationProvider() db_types.MigrationsProvider {
	// return singleton
	return p
}

func NewMigrationProviderModule() fx.Option {
	return util.NewFxModule[db_types.MigrationsProvider](
		"realworld_app_migrations_provider",
		NewMigrationProvider,
	)
}

func (m *migrationsProviderImpl) GetSqlMigrationsFs() fs.FS {
	return sqlMigrations
}

func (m *migrationsProviderImpl) GetCodeMigrations() *[]db_types.GoMigration {
	return codeMigrations
}

// registerCodeMigration registers a code migration to be included in the migrations list,
// you MUST include this in an init() function in any go file that defines a code based migration
func registerCodeMigration(migration db_types.GoMigration) {
	lock.Lock()
	defer lock.Unlock()
	*codeMigrations = append(*codeMigrations, migration)
}
