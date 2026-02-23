package user_types

import (
	"context"

	"github.com/google/uuid"
)

//mockery:generate: true
type UserValidations interface {
	ValidateUser(user User) DomainError
	ValidateUsernameDoesNotConflict(ctx context.Context, username string) DomainError
	ValidateEmailDoesNotConflict(ctx context.Context, email string) DomainError
	ValidateUserIdExists(ctx context.Context, id uuid.UUID) (User, DomainError)
	ValidateUsernameExists(ctx context.Context, email string) (User, DomainError)
	ValidateCanFollow(followedByUserId, followingUserId uuid.UUID) DomainError
}
