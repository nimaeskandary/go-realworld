package routes

import (
	"context"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/internal/transformers"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/context"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"
	"golang.org/x/sync/errgroup"
)

type ProfileRoutes struct {
	userService user_types.UserService
}

func NewProfileRoutes(
	userService user_types.UserService,
) *ProfileRoutes {
	return &ProfileRoutes{
		userService: userService,
	}
}

func (r *ProfileRoutes) GetProfileByUsername(ctx context.Context, request api_gen.GetProfileByUsernameRequestObject) (api_gen.GetProfileByUsernameResponseObject, error) {
	// auth is optional for this route
	authUser := auth_context.UserFromCtx(ctx)

	user := user_types.User{}
	isFollowing := false

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		user, err = r.userService.GetUserByUsername(egCtx, request.Username)
		return err
	})
	eg.Go(func() error {
		if authUser.IsNone() {
			return nil
		}
		var err error
		isFollowing, err = r.userService.IsFollowing(egCtx, authUser.MustGet(), request.Username)
		return err
	})
	err := eg.Wait()

	if err != nil {
		switch user_types.AsDomainError(err).(type) {
		case user_types.NotFoundError:
			return api_gen.GetProfileByUsername422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(err),
			}, nil
		default:
			return nil, err
		}
	}

	return api_gen.GetProfileByUsername200JSONResponse{ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{Profile: transformers.ToApiProfile(user, isFollowing)}}, nil
}

func (r *ProfileRoutes) UnfollowUserByUsername(ctx context.Context, request api_gen.UnfollowUserByUsernameRequestObject) (api_gen.UnfollowUserByUsernameResponseObject, error) {
	authUser := auth_context.UserFromCtx(ctx)
	if authUser.IsNone() {
		return api_gen.UnauthorizedResponse{}, nil
	}

	user, err := r.userService.UnfollowProfile(ctx, authUser.MustGet(), request.Username)
	if err != nil {
		switch user_types.DomainError(err).(type) {
		case user_types.NotFoundError:
			return api_gen.UnfollowUserByUsername422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(err),
			}, nil
		default:
			return nil, err
		}
	}

	return api_gen.UnfollowUserByUsername200JSONResponse{ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{Profile: transformers.ToApiProfile(user, false)}}, nil
}

func (r *ProfileRoutes) FollowUserByUsername(ctx context.Context, request api_gen.FollowUserByUsernameRequestObject) (api_gen.FollowUserByUsernameResponseObject, error) {
	authUser := auth_context.UserFromCtx(ctx)
	if authUser.IsNone() {
		return api_gen.UnauthorizedResponse{}, nil
	}

	user, err := r.userService.FollowProfile(ctx, authUser.MustGet(), request.Username)
	if err != nil {
		switch user_types.DomainError(err).(type) {
		case user_types.NotFoundError,
			user_types.CannotFollowYourselfError:
			return api_gen.FollowUserByUsername422JSONResponse{
				GenericErrorJSONResponse: transformers.ToApiError(err),
			}, nil
		default:
			return nil, err
		}
	}

	return api_gen.FollowUserByUsername200JSONResponse{ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{Profile: transformers.ToApiProfile(user, true)}}, nil
}
