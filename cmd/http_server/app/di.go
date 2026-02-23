package app

import (
	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler"
	"github.com/nimaeskandary/go-realworld/pkg/article"
	"github.com/nimaeskandary/go-realworld/pkg/auth"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/config"
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	"github.com/nimaeskandary/go-realworld/pkg/database"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	obs "github.com/nimaeskandary/go-realworld/pkg/observability"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/nimaeskandary/go-realworld/pkg/user"

	"go.uber.org/fx"
)

func ModuleList(configData []byte) []fx.Option {
	return []fx.Option{
		config.NewIdentitySecretParserModule(),
		config.NewYamlConfigLoaderModule[Config](configData),
		fx.Provide(
			func(cfg config_types.ConfigLoader[Config]) HttpServerConfig {
				return cfg.GetConfig().HttpServer
			},
			func(cfg config_types.ConfigLoader[Config]) auth_types.JwtAuthServiceConfig {
				return cfg.GetConfig().JwtAuthService
			},
			func(cfg config_types.ConfigLoader[Config]) obs_types.SlogLoggerConfig {
				return cfg.GetConfig().Slog
			},
			func(cfg config_types.ConfigLoader[Config]) db_types.RealWorldAppDbConfig {
				return cfg.GetConfig().RealWorldAppDb
			},
		),
		database.NewPostgresRealworldAppDbModule[db_types.PostgresRealWorldAppDb](),
		http_handler.NewHttpHandlerModule(),
		auth.NewAuthModule(),
		user.NewUserModule(),
		article.NewArticleModule(),
		obs.NewSlogLoggerModule(),
	}
}
