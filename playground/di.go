package playground

import (
	"context"
	"log"
	"os"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler"
	"github.com/nimaeskandary/go-realworld/pkg/article"
	article_types "github.com/nimaeskandary/go-realworld/pkg/article/types"
	"github.com/nimaeskandary/go-realworld/pkg/auth"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/config"
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	"github.com/nimaeskandary/go-realworld/pkg/database"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	obs "github.com/nimaeskandary/go-realworld/pkg/observability"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/nimaeskandary/go-realworld/pkg/user"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

type Config struct {
	JwtAuthService auth_types.JwtAuthServiceConfig `json:"jwt_auth_service" validate:"required"`
	Slog           obs_types.SlogLoggerConfig      `json:"slog" validate:"required"`
	RealWorldAppDb db_types.RealWorldAppDbConfig   `json:"realworld_app_db" validate:"required"`
}

type StandardSystem struct {
	ArticleRepo    article_types.ArticleRepository
	ArticleService article_types.ArticleService
	AuthService    auth_types.AuthService
	UserRepo       user_types.UserRepository
	UserService    user_types.UserService
}

func SetupStandardSystem(ctx context.Context) (StandardSystem, util.CleanupManager) {
	configData, err := os.ReadFile("config/local.yaml")
	if err != nil {
		log.Fatalf("failed to read config file %v", err)
	}

	f := StandardSystem{}
	app := util.CreateFxAppAndExtract(ModuleList(configData),
		&f.ArticleRepo,
		&f.ArticleService,
		&f.AuthService,
		&f.UserRepo,
		&f.UserService,
	)

	err = app.Start(ctx)
	if err != nil {
		log.Fatalf("failed to start fx app: %v", err)
	}

	cm := util.NewCleanupManager(ctx, true)
	cm.RegisterCleanupFunc(func() {
		log.Printf("stopping fx app...")
		err := app.Stop(ctx)
		if err != nil {
			log.Printf("failed to stop fx app: %v", err)
		}
	})

	return f, cm
}

func ModuleList(configData []byte) []fx.Option {
	return []fx.Option{
		config.NewIdentitySecretParserModule(),
		config.NewYamlConfigLoaderModule[Config](configData),
		fx.Provide(
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
