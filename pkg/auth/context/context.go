package auth_context

import (
	"context"

	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/samber/mo"
)

type userIdKey struct{}

func CtxWithUser(ctx context.Context, user user_types.User) context.Context {
	return context.WithValue(ctx, userIdKey{}, user)
}

func UserFromCtx(ctx context.Context) mo.Option[user_types.User] {
	user, ok := ctx.Value(userIdKey{}).(user_types.User)
	if !ok {
		return mo.None[user_types.User]()
	}
	return mo.Some(user)
}
