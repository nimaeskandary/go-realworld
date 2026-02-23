package transformers

import (
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/samber/mo"
)

func ToApiUser(user user_types.User, token string) api_gen.User {
	return api_gen.User{
		Username: user.Username,
		Email:    user.Email,
		Bio:      user.Bio.OrElse(""),
		Image:    user.Image.OrElse(""),
		Token:    token,
	}
}

func FromApiUpdateUser(fromApi api_gen.UpdateUser, currentUser user_types.User) user_types.UpsertUserParams {
	return user_types.UpsertUserParams{
		Username: func() string {
			if fromApi.Username != nil {
				return *fromApi.Username
			}
			return currentUser.Username
		}(),
		Email: func() string {
			if fromApi.Email != nil {
				return *fromApi.Email
			}
			return currentUser.Email
		}(),
		Bio: func() mo.Option[string] {
			if fromApi.Bio != nil {
				return mo.Some(*fromApi.Bio)
			}
			return currentUser.Bio
		}(),
		Image: func() mo.Option[string] {
			if fromApi.Image != nil {
				return mo.Some(*fromApi.Image)
			}
			return currentUser.Image
		}(),
	}
}

func FromApiNewUser(fromApi api_gen.NewUser) user_types.UpsertUserParams {
	return user_types.UpsertUserParams{
		Username: fromApi.Username,
		Email:    fromApi.Email,
		Token:    mo.None[string](),
		Bio:      mo.None[string](),
		Image:    mo.None[string](),
	}
}
