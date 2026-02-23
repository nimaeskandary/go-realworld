package user_types

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

type UserService interface {
	CreateUser(ctx context.Context, user UpsertUserParams) (User, DomainError)
	UpdateUser(ctx context.Context, id uuid.UUID, updated UpsertUserParams) (User, DomainError)
	DeleteUser(ctx context.Context, id uuid.UUID) DomainError
	GetUserByEmail(ctx context.Context, email string) (User, DomainError)
	GetUserByUsername(ctx context.Context, username string) (User, DomainError)
	IsFollowing(ctx context.Context, authUser User, targetUsername string) (bool, DomainError)
	FollowProfile(ctx context.Context, authUser User, targetUsername string) (User, DomainError)
	UnfollowProfile(ctx context.Context, authUser User, targetUsername string) (User, DomainError)
}

type UpsertUserParams struct {
	Username string
	Email    string
	Token    mo.Option[string]
	Bio      mo.Option[string]
	Image    mo.Option[string]
}
