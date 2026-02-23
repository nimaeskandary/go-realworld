package internal

import (
	"context"
	"time"

	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type userServiceImpl struct {
	userRepo    user_types.UserRepository
	validations user_types.UserValidations
}

func NewUserServiceImpl(userRepo user_types.UserRepository, validations user_types.UserValidations) user_types.UserService {
	return &userServiceImpl{
		userRepo:    userRepo,
		validations: validations,
	}
}

func (s *userServiceImpl) CreateUser(ctx context.Context, params user_types.UpsertUserParams) (user_types.User, user_types.DomainError) {
	now := time.Now().UnixMilli()
	newUser := user_types.User{
		Id:              uuid.New(),
		Username:        params.Username,
		Email:           params.Email,
		Bio:             params.Bio,
		Image:           params.Image,
		CreatedAtMillis: now,
		UpdatedAtMillis: now,
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error { return s.validations.ValidateUser(newUser) })
	eg.Go(func() error { return s.validations.ValidateUsernameDoesNotConflict(egCtx, newUser.Username) })
	eg.Go(func() error { return s.validations.ValidateEmailDoesNotConflict(egCtx, newUser.Email) })

	if err := eg.Wait(); err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	created, err := s.userRepo.UpsertUser(ctx, newUser)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	return created, nil
}

func (s *userServiceImpl) UpdateUser(ctx context.Context, id uuid.UUID, params user_types.UpsertUserParams) (user_types.User, user_types.DomainError) {
	var err error

	existingUser, err := s.validations.ValidateUserIdExists(ctx, id)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	updatedUser := user_types.User{
		Id:              existingUser.Id,
		Username:        params.Username,
		Email:           params.Email,
		Bio:             params.Bio,
		Image:           params.Image,
		CreatedAtMillis: existingUser.CreatedAtMillis,
		UpdatedAtMillis: time.Now().UnixMilli(),
	}

	err = s.validations.ValidateUser(updatedUser)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	eg, egCtx := errgroup.WithContext(ctx)

	if updatedUser.Username != existingUser.Username {
		eg.Go(func() error { return s.validations.ValidateUsernameDoesNotConflict(egCtx, updatedUser.Username) })
	}
	if updatedUser.Email != existingUser.Email {
		eg.Go(func() error { return s.validations.ValidateEmailDoesNotConflict(egCtx, updatedUser.Email) })
	}

	if err := eg.Wait(); err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	updated, err := s.userRepo.UpsertUser(ctx, updatedUser)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	return updated, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID) user_types.DomainError {
	err := s.userRepo.DeleteUser(ctx, id)

	if err != nil {
		return user_types.AsDomainError(err)
	}

	return nil
}

func (s *userServiceImpl) GetUserByEmail(ctx context.Context, email string) (user_types.User, user_types.DomainError) {
	result, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	if result.IsNone() {
		return user_types.User{}, user_types.NotFoundError{Identifier: email}
	}
	return result.MustGet(), nil
}

func (s *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (user_types.User, user_types.DomainError) {
	result, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	if result.IsNone() {
		return user_types.User{}, user_types.NotFoundError{Identifier: username}
	}
	return result.MustGet(), nil
}

func (s *userServiceImpl) IsFollowing(ctx context.Context, authUser user_types.User, targetUsername string) (bool, user_types.DomainError) {
	var err error
	targetUser, err := s.validations.ValidateUsernameExists(ctx, targetUsername)
	if err != nil {
		return false, user_types.AsDomainError(err)
	}

	isFollowing, err := s.userRepo.IsFollowing(ctx, authUser.Id, targetUser.Id)
	if err != nil {
		return false, user_types.AsDomainError(err)
	}
	return isFollowing, nil
}

func (s *userServiceImpl) FollowProfile(ctx context.Context, authUser user_types.User, targetUsername string) (user_types.User, user_types.DomainError) {
	var err error
	targetUser, err := s.validations.ValidateUsernameExists(ctx, targetUsername)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	err = s.validations.ValidateCanFollow(authUser.Id, targetUser.Id)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	err = s.userRepo.Follow(ctx, authUser.Id, targetUser.Id)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	return targetUser, nil
}

func (s *userServiceImpl) UnfollowProfile(ctx context.Context, authUser user_types.User, targetUsername string) (user_types.User, user_types.DomainError) {
	var err error
	targetUser, err := s.validations.ValidateUsernameExists(ctx, targetUsername)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}

	err = s.userRepo.Unfollow(ctx, authUser.Id, targetUser.Id)
	if err != nil {
		return user_types.User{}, user_types.AsDomainError(err)
	}
	return targetUser, nil
}
