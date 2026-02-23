package app

import (
	"fmt"

	"github.com/nimaeskandary/go-realworld/pkg/config"
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	"github.com/nimaeskandary/go-realworld/pkg/database"
	"github.com/nimaeskandary/go-realworld/pkg/database/migrations/realworld_app"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"

	"go.uber.org/fx"
)

func realworldAppDbModules(configData []byte) []fx.Option {
	return []fx.Option{
		config.NewIdentitySecretParserModule(),
		config.NewYamlConfigLoaderModule[Config](configData),
		fx.Provide(
			func(cfg config_types.ConfigLoader[Config]) db_types.RealWorldAppDbConfig {
				return cfg.GetConfig().RealWorldAppDb
			},
		),
		database.NewPostgresRealworldAppDbModule[db_types.SQLDatabase](),
		realworld_app.NewMigrationProviderModule(),
		database.NewGooseMigrationRunnerModule(),
	}
}

func ModuleList(targetDatabase string, configData []byte) ([]fx.Option, error) {
	switch targetDatabase {
	case "realworld_app":
		return realworldAppDbModules(configData), nil
	default:
		return nil, fmt.Errorf("unknown target database: %v", targetDatabase)
	}
}
