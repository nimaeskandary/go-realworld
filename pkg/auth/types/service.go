package auth_types

import (
	"context"
)

//mockery:generate: true
type AuthService interface {
	GenerateToken(ctx context.Context, username string) (AuthToken, DomainError)
	ParseToken(ctx context.Context, token string) (AuthToken, DomainError)
}
