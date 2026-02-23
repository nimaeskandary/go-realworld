package auth

import (
	"github.com/nimaeskandary/go-realworld/pkg/auth/internal"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewAuthModule() fx.Option {
	return util.NewFxModule[auth_types.AuthService](
		"auth_service",
		internal.NewJWTAuthService,
	)
}
