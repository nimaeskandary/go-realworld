package internal

import (
	"net/http"

	http_handler_types "github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/types"
	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/middleware"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	obstypes "github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

type httpHandlerImpl struct {
	handler http.Handler
}

func NewHttpHandlerImpl(
	generateRoutesImpl api_gen.StrictServerInterface,
	authService auth_types.AuthService,
	logger obstypes.Logger,
) http_handler_types.HttpHandler {
	// logically these are run in reverse order due to wrapping, the first element here is the outermost middleware
	middlewares := []api_gen.StrictMiddlewareFunc{
		middleware.AuthenticateRoute(logger),
		middleware.CreateAuthContext(logger, authService),
		middleware.HandleError(logger),
	}

	handler := api_gen.Handler(api_gen.NewStrictHandler(generateRoutesImpl, middlewares))

	return &httpHandlerImpl{
		handler: handler,
	}
}

func (h *httpHandlerImpl) GetHandler() http.Handler {
	return h.handler
}
