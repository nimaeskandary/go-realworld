package routes

import (
	"context"
	"fmt"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/internal/transformers"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/context"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
)

type UserRoutes struct {
	authService auth_types.AuthService
	userService user_types.UserService
}

func NewUserRoutes(
	authService auth_types.AuthService,
	userService user_types.UserService,
) *UserRoutes {
	return &UserRoutes{
		authService: authService,
		userService: userService,
	}
}

func (r *UserRoutes) GetCurrentUser(ctx context.Context, request api_gen.GetCurrentUserRequestObject) (api_gen.GetCurrentUserResponseObject, error) {
	authUser := auth_context.UserFromCtx(ctx)
	if authUser.IsNone() {
		return api_gen.UnauthorizedResponse{}, nil
	}

	token, err := r.authService.GenerateToken(ctx, authUser.MustGet().Username)
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %w", err)
	}

	return api_gen.GetCurrentUser200JSONResponse{
		UserResponseJSONResponse: api_gen.UserResponseJSONResponse{
			User: transformers.ToApiUser(authUser.MustGet(), token.GetTokenString()),
		},
	}, nil
}

func (r *UserRoutes) UpdateCurrentUser(ctx context.Context, request api_gen.UpdateCurrentUserRequestObject) (api_gen.UpdateCurrentUserResponseObject, error) {
	authUser := auth_context.UserFromCtx(ctx)
	if authUser.IsNone() {
		return api_gen.UnauthorizedResponse{}, nil
	}
	var err error
	updatedUser, err := r.userService.UpdateUser(ctx, authUser.MustGet().Id, transformers.FromApiUpdateUser(request.Body.User, authUser.MustGet()))

	if err != nil {
		switch user_types.DomainError(user_types.AsDomainError(err)).(type) {
		case
			user_types.ConflictError,
			user_types.NotFoundError,
			user_types.BadParamsError:
			return api_gen.UpdateCurrentUser422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(err),
			}, nil
		default:
			return nil, fmt.Errorf("error updating user: %w", err)
		}
	}

	token, err := r.authService.GenerateToken(ctx, updatedUser.Username)
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %w", err)
	}

	return api_gen.UpdateCurrentUser200JSONResponse{
		UserResponseJSONResponse: api_gen.UserResponseJSONResponse{
			User: transformers.ToApiUser(updatedUser, token.GetTokenString()),
		},
	}, nil
}

func (r *UserRoutes) CreateUser(ctx context.Context, request api_gen.CreateUserRequestObject) (api_gen.CreateUserResponseObject, error) {
	var err error
	created, err := r.userService.CreateUser(ctx, transformers.FromApiNewUser(request.Body.User))
	if err != nil {
		switch user_types.DomainError(user_types.AsDomainError(err)).(type) {
		case
			user_types.ConflictError,
			user_types.BadParamsError:
			return api_gen.CreateUser422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(err),
			}, nil
		default:
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	token, err := r.authService.GenerateToken(ctx, created.Username)
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %w", err)
	}

	return api_gen.CreateUser201JSONResponse{
		UserResponseJSONResponse: api_gen.UserResponseJSONResponse{
			User: transformers.ToApiUser(created, token.GetTokenString()),
		},
	}, nil
}

func (r *UserRoutes) Login(ctx context.Context, request api_gen.LoginRequestObject) (api_gen.LoginResponseObject, error) {
	// just check user exists, no password validation
	var err error
	user, err := r.userService.GetUserByEmail(ctx, request.Body.User.Email)

	if err != nil {
		switch user_types.DomainError(user_types.AsDomainError(err)).(type) {
		case
			user_types.NotFoundError:
			return api_gen.Login422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(fmt.Errorf("user not found")),
			}, nil
		default:
			return nil, fmt.Errorf("error logging in user: %w", err)
		}
	}

	token, err := r.authService.GenerateToken(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("error generating auth token: %w", err)
	}

	return api_gen.Login200JSONResponse{
		UserResponseJSONResponse: api_gen.UserResponseJSONResponse{
			User: transformers.ToApiUser(user, token.GetTokenString()),
		},
	}, nil
}
