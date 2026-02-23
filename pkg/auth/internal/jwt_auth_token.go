package internal

import (
	"time"

	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtClaims struct {
	UserId   uuid.UUID `json:"id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

type jwtToken struct {
	claims      *jwtClaims
	tokenString string
	user        user_types.User
}

func NewJwtToken(tokenString string, claims *jwtClaims, user user_types.User) auth_types.AuthToken {
	return &jwtToken{
		tokenString: tokenString,
		claims:      claims,
		user:        user,
	}
}

func (t *jwtToken) GetTokenString() string {
	return t.tokenString
}

func (t *jwtToken) GetUser() user_types.User {
	return t.user
}

func (t *jwtToken) GetTTL() time.Duration {
	if t.claims.ExpiresAt == nil {
		return 0
	}
	return time.Until(t.claims.ExpiresAt.Time)
}

func (t *jwtToken) IsExpired() bool {
	if t.claims.ExpiresAt == nil {
		return true
	}
	return time.Now().After(t.claims.ExpiresAt.Time)
}
