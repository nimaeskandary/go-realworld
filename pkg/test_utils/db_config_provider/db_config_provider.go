package db_config_provider

import (
	"context"

	"github.com/nimaeskandary/go-realworld/pkg/database"
	"github.com/nimaeskandary/go-realworld/pkg/database/migrations/realworld_app"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
)

// SqlDbConfigProvider - interface for providing SQL database configs, with the ability to cleanup after use.
// The intention here, is to implement providers that create isolated databases on each GetFreshDbConfig call and return its config,
// this allows for integration tests that rely on databases to run in parallel
type SqlDbConfigProvider interface {
	// GetFreshDbConfig - returns a config for a new, fresh database.
	// The provider implementation is responsible for creating the database and running any necessary migrations on it before returning the config
	GetFreshDbConfig(ctx context.Context) (db_types.SqlDbConfig, error)
	// Cleanup - cleans up any resources used by the provider for the given db config, e.g. dropping the database
	Cleanup(ctx context.Context, dbconfig db_types.SqlDbConfig) error
}

// RealWorldAppDbConfigProvider - provides configs for a realworld app db
func RealWorldAppDbConfigProvider() SqlDbConfigProvider {
	// config to hit test postgres docker container
	cfg := db_types.RealWorldAppDbConfig{
		Host:     "localhost",
		Port:     15432,
		Username: "testpostgres",
		Password: "testpassword",
		DBName:   "realworld_app",
		SslMode:  "disable",
	}

	return NewPostgresSqlDbConfigProvider(
		db_types.SqlDbConfig(cfg),
		func(ctx context.Context, db db_types.SQLDatabase) error {
			migrationRunner, err := database.NewGooseMigrationRunner(db, realworld_app.NewMigrationProvider())
			if err != nil {
				return err
			}
			err = migrationRunner.Start(ctx)
			if err != nil {
				return err
			}
			defer func() {
				_ = migrationRunner.Stop(ctx)
			}()
			return migrationRunner.ApplyAll(ctx)
		},
	)
}
