package auth_types

import (
	"time"

	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
)

//mockery:generate: true
type AuthToken interface {
	GetTokenString() string
	GetUser() user_types.User
	GetTTL() time.Duration
	IsExpired() bool
}
