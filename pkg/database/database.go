package database

import (
	"github.com/nimaeskandary/go-realworld/pkg/database/internal"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

// NewRealworldAppDbModule provides the RealWorld application database module, this can be casted to help with
// running migrations, as that expects there to just be one generic SQLDatabase type loaded
func NewPostgresRealworldAppDbModule[T db_types.SQLDatabase]() fx.Option {
	return util.NewFxModuleWithLifecycle[T]("realworld_app_db",
		func(cfg db_types.RealWorldAppDbConfig) (db_types.SQLDatabase, error) {
			return internal.NewPostgresSQLDatabase(db_types.SqlDbConfig(cfg))
		})
}

func NewGooseMigrationRunnerModule() fx.Option {
	return util.NewFxModuleWithLifecycle[db_types.SqlMigrationRunner](
		"goose_migration_runner",
		internal.NewGooseMigrationRunner,
	)
}

var NewPostgresSQLDatabase = internal.NewPostgresSQLDatabase
var NewGooseMigrationRunner = internal.NewGooseMigrationRunner
