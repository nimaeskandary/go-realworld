package internal

import (
	"context"

	"github.com/nimaeskandary/go-realworld/pkg/user/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type userValidationsImpl struct {
	validator *validator.Validate
	userRepo  user_types.UserRepository
}

func NewUserValidationsImpl(userRepo user_types.UserRepository) user_types.UserValidations {
	return &userValidationsImpl{
		validator: util.NewValidator(),
		userRepo:  userRepo,
	}
}

func (v *userValidationsImpl) ValidateUser(user user_types.User) user_types.DomainError {
	err := v.validator.Struct(user)
	if err != nil {
		return user_types.BadParamsError{Err: err}
	}
	return nil
}

func (v *userValidationsImpl) ValidateUsernameDoesNotConflict(ctx context.Context, username string) user_types.DomainError {
	opt, err := v.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return user_types.AsDomainError(err)
	}
	if opt.IsSome() {
		return user_types.ConflictError{
			Msg: "username already exists",
		}
	}
	return nil
}

func (v *userValidationsImpl) ValidateEmailDoesNotConflict(ctx context.Context, email string) user_types.DomainError {
	opt, err := v.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return user_types.AsDomainError(err)
	}
	if opt.IsSome() {
		return user_types.ConflictError{
			Msg: "email already exists",
		}
	}
	return nil
}

func (v *userValidationsImpl) ValidateUserIdExists(ctx context.Context, id uuid.UUID) (user_types.User, user_types.DomainError) {
	existingUserOpt, err := v.userRepo.GetUserById(ctx, id)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	existingUser, ok := existingUserOpt.Get()

	if !ok {
		return user_types.User{}, user_types.NotFoundError{Identifier: id.String()}
	}

	return existingUser, nil
}

func (v *userValidationsImpl) ValidateUsernameExists(ctx context.Context, username string) (user_types.User, user_types.DomainError) {
	existingUserOpt, err := v.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	existingUser, ok := existingUserOpt.Get()

	if !ok {
		return user_types.User{}, user_types.NotFoundError{Identifier: username}
	}

	return existingUser, nil
}

func (v *userValidationsImpl) ValidateCanFollow(followedByUserId, followingUserId uuid.UUID) user_types.DomainError {
	if followedByUserId.String() == followingUserId.String() {
		return user_types.CannotFollowYourselfError{}
	}
	return nil
}
