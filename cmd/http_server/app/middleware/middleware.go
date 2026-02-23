package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

// these are defined as operation_id for each route in the OpenAPI spec, can be used to conditionally apply middleware logic
// based on the route being accessed
type operationId string

const (
	LoginOpId                operationId = "Login"
	CreateUserOpId           operationId = "CreateUser"
	GetCurrentUserOpId       operationId = "GetCurrentUser"
	GetProfileByUsernameOpId operationId = "GetProfileByUsername"
)

var errUnexpected = fmt.Errorf("unexpected error occured")

// WithCorsMiddleware allows cross-origin requests from the specified allowed origins, otherwise blocks the request with a 403
func WithCorsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// handle preflight requests and just return 204
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	})
}

// HandleError is a middleware that intercepts unexpected errors from handlers and logs them,
// then returns a generic message to the client to obfuscate internal details.
func HandleError(logger obs_types.Logger) api_gen.StrictMiddlewareFunc {
	return func(next api_gen.StrictHandlerFunc, operationID string) api_gen.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (res any, err error) {
			defer func() {
				if rec := recover(); rec != nil {
					// TODO add trace id to message if available in context
					logger.Error(ctx, "panic recovered in middleware", fmt.Errorf("panic: %v", rec), "operation_id", operationID)
					res = nil
					err = errUnexpected
				}
			}()

			// run the rest of the route first
			resp, err := next(ctx, w, r, request)

			if err != nil {
				logger.Error(ctx, "unexpected error handling request", err, "operation_id", operationID)
				// TODO add trace id to message if available in context
				return resp, errUnexpected
			}

			return resp, nil
		}
	}
}
