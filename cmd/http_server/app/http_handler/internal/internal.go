package internal

import (
	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/internal/routes"

	"go.uber.org/fx"
)

func RoutesModule() fx.Option {
	return fx.Provide(
		fx.Private,
		NewGeneratedRoutesImpl,
		routes.NewUserRoutes,
		routes.NewProfileRoutes,
	)
}
