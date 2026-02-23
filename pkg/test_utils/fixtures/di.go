package fixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler"
	"github.com/nimaeskandary/go-realworld/pkg/article"
	"github.com/nimaeskandary/go-realworld/pkg/auth"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/database"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	obs "github.com/nimaeskandary/go-realworld/pkg/observability"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/config"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/db_config_provider"
	"github.com/nimaeskandary/go-realworld/pkg/user"

	"go.uber.org/fx"
)

func AllTestModules(t *testing.T) ([]fx.Option, error) {
	testModules := TestModules()
	testDatabaseModules, err := TestDatabaseModules(t)
	if err != nil {
		return nil, fmt.Errorf("failed to create test database modules: %w", err)
	}
	return append(testModules, testDatabaseModules...), nil
}

func TestModules() []fx.Option {
	return []fx.Option{
		fx.Provide(
			func() config.Config {
				return config.NewTestConfig()
			},
			func(c config.Config) obs_types.SlogLoggerConfig {
				return c.Slog
			},
			func(c config.Config) auth_types.JwtAuthServiceConfig {
				return c.JwtAuthService
			},
		),
		auth.NewAuthModule(),
		http_handler.NewHttpHandlerModule(),
		obs.NewSlogLoggerModule(),
		user.NewUserModule(),
		article.NewArticleModule(),
	}
}

// TestDatabaseModules returns fx options for test databases. Note, isolated databases are made on call, so recommend setting this up in fixtures at the top level,
// and not per test case with the intention of parallelizing each sub test, as that may negatively impact test performance more than any parallization will gain.
func TestDatabaseModules(t *testing.T) ([]fx.Option, error) {
	realWorldAppDb, err := NewRealWorldAppDbTestModule(t)
	if err != nil {
		return nil, fmt.Errorf("failed to create realworld app db module: %w", err)
	}

	return []fx.Option{
		realWorldAppDb,
	}, nil
}

// NewRealWorldAppDbTestModule - fx module for a realworld app db that can be used in tests
func NewRealWorldAppDbTestModule(t *testing.T) (fx.Option, error) {
	realWorldDatabaseProvider := db_config_provider.RealWorldAppDbConfigProvider()
	realWorldDbCfg, err := realWorldDatabaseProvider.GetFreshDbConfig(t.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get realworld db config: %w", err)
	}

	t.Cleanup(func() { _ = realWorldDatabaseProvider.Cleanup(context.Background(), realWorldDbCfg) })

	return fx.Module("test_postgres_realworld_app_db",
		fx.Provide(
			func() db_types.RealWorldAppDbConfig {
				return db_types.RealWorldAppDbConfig(realWorldDbCfg)
			},
		),
		database.NewPostgresRealworldAppDbModule[db_types.PostgresRealWorldAppDb](),
	), nil
}
