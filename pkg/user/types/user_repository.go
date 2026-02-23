package user_types

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

//mockery:generate: true
type UserRepository interface {
	UpsertUser(ctx context.Context, user User) (User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUserByUsername(ctx context.Context, username string) (mo.Option[User], error)
	GetUserById(ctx context.Context, id uuid.UUID) (mo.Option[User], error)
	GetUserByEmail(ctx context.Context, email string) (mo.Option[User], error)
	IsFollowing(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) (bool, error)
	Follow(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) error
	Unfollow(ctx context.Context, followedByUserId uuid.UUID, followingUserId uuid.UUID) error
}
