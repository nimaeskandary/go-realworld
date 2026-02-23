package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/context"
	"github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

var nonAuthenticatedOperations = []operationId{
	LoginOpId,
	CreateUserOpId,
	GetProfileByUsernameOpId,
}

// CreateAuthContext is a middleware that gets the auth header if it exists, and adds the auth user to the context.
// the OpenApi spec specifies the security schema 'Authorization': 'Token xxxxxx.yyyyyyy.zzzzzz'.
// This is separate from AuthenticateRoute, because some api routes allow for optional authentication, where we just
// want to know the auth user if it is provided
func CreateAuthContext(logger obs_types.Logger, authService auth_types.AuthService) api_gen.StrictMiddlewareFunc {
	return func(next api_gen.StrictHandlerFunc, opId string) api_gen.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Token ")
			if token == "" {
				return next(ctx, w, r, request)
			}

			parsed, err := authService.ParseToken(ctx, token)
			if err != nil {
				logger.Error(ctx, "error parsing auth token", err, "token", token)
				return api_gen.UnauthorizedResponse{}, nil
			}

			ctx = logger.CtxWithLogAttributes(ctx, "user_id", parsed.GetUser().Id.String(), "user_username", parsed.GetUser().Username)
			ctx = auth_context.CtxWithUser(ctx, parsed.GetUser())

			return next(ctx, w, r, request)
		}
	}
}

// AuthenticateRoute is a middleware that blocks unauthenticated requests. It must be called after CreateAuthContext in the middleware chain.
func AuthenticateRoute(logger obs_types.Logger) api_gen.StrictMiddlewareFunc {
	return func(next api_gen.StrictHandlerFunc, opId string) api_gen.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			// ignore non authenticated routes
			if slices.Contains(nonAuthenticatedOperations, operationId(opId)) {
				return next(ctx, w, r, request)
			}

			authUser := auth_context.UserFromCtx(ctx)
			if authUser.IsNone() {
				return api_gen.UnauthorizedResponse{}, nil
			}

			return next(ctx, w, r, request)
		}
	}
}
