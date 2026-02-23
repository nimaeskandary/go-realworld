package user

import (
	"github.com/nimaeskandary/go-realworld/pkg/user/internal"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewUserModule() fx.Option {
	return util.NewFxModule[user_types.UserService](
		"user_service",
		internal.NewUserServiceImpl,
		fx.Provide(
			internal.NewPostgresUserRepository,
			internal.NewUserValidationsImpl,
		),
	)
}
