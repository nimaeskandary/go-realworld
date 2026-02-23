package transformers

import (
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"
)

func ToApiProfile(user user_types.User, isFollowing bool) api_gen.Profile {
	return api_gen.Profile{
		Username:  user.Username,
		Bio:       user.Bio.OrElse(""),
		Image:     user.Image.OrElse(""),
		Following: isFollowing,
	}
}
