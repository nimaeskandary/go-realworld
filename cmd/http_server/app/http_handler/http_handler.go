package http_handler

import (
	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/internal"
	http_handler_types "github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewHttpHandlerModule() fx.Option {
	return util.NewFxModule[http_handler_types.HttpHandler](
		"api",
		internal.NewHttpHandlerImpl,
		internal.RoutesModule(),
	)
}
