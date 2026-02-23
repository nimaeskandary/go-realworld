package internal_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"
	"github.com/nimaeskandary/go-realworld/pkg/user/internal"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
	user_types_mocks "github.com/nimaeskandary/go-realworld/pkg/user/types/mocks"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_UserServiceImpl(t *testing.T) {
	t.Parallel()

	type testFixture struct {
		userRepoMock    *user_types_mocks.MockUserRepository
		validationsMock *user_types_mocks.MockUserValidations
		underTest       user_types.UserService
	}

	setup := func(t *testing.T) testFixture {
		userRepoMock := user_types_mocks.NewMockUserRepository(t)
		validationsMock := user_types_mocks.NewMockUserValidations(t)
		underTest := internal.NewUserServiceImpl(userRepoMock, validationsMock)

		return testFixture{
			userRepoMock:    userRepoMock,
			validationsMock: validationsMock,
			underTest:       underTest,
		}
	}

	t.Run("CreateUser", func(t *testing.T) {
		t.Parallel()

		params := user_types.UpsertUserParams{
			Username: "testuser",
			Email:    "testuser@example.com",
		}

		t.Run("should fail if ValidateUser fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(user_types.BadParamsError{})
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)

			created, err := f.underTest.CreateUser(t.Context(), params)

			assert.Empty(t, created)
			assert.IsType(t, user_types.BadParamsError{}, err)
		})

		t.Run("fail if ValidateUsernameDoesNotConflict fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(user_types.ConflictError{})
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)

			created, err := f.underTest.CreateUser(t.Context(), params)

			assert.Empty(t, created)
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("fail if ValidateEmailDoesNotConflict fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(user_types.ConflictError{})

			created, err := f.underTest.CreateUser(t.Context(), params)

			assert.Empty(t, created)
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("should fail with UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			expectedUser := helpers.GenUserWithUpsertParams(params)

			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)

			f.userRepoMock.EXPECT().UpsertUser(
				mock.Anything,
				mock.MatchedBy(helpers.UserMatchesUpsertParams(params)),
			).Return(expectedUser, errors.New("query failure"))

			created, err := f.underTest.CreateUser(t.Context(), params)
			assert.Empty(t, created)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should create user successfully", func(t *testing.T) {
			t.Parallel()

			t.Run("without optional fields", func(t *testing.T) {
				t.Parallel()
				f := setup(t)

				upsertParams := params
				upsertParams.Token = mo.None[string]()
				upsertParams.Bio = mo.None[string]()
				upsertParams.Image = mo.None[string]()

				expectedUser := helpers.GenUserWithUpsertParams(upsertParams)

				f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).Return(nil)
				f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, upsertParams.Username).Return(nil)
				f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, upsertParams.Email).Return(nil)

				f.userRepoMock.EXPECT().
					UpsertUser(
						mock.Anything,
						mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams)),
					).
					Return(expectedUser, nil)

				created, err := f.underTest.CreateUser(t.Context(), upsertParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedUser, created)
			})

			t.Run("with optional fields", func(t *testing.T) {
				t.Parallel()
				f := setup(t)
				upsertParams := params
				upsertParams.Token = mo.Some("test-token")
				upsertParams.Bio = mo.Some("test bio")
				upsertParams.Image = mo.Some("http://example.com/image.png")

				expectedUser := helpers.GenUserWithUpsertParams(upsertParams)

				f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).Return(nil)
				f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, upsertParams.Username).Return(nil)
				f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, upsertParams.Email).Return(nil)

				f.userRepoMock.EXPECT().
					UpsertUser(
						mock.Anything,
						mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams)),
					).
					Return(expectedUser, nil)

				created, err := f.underTest.CreateUser(t.Context(), upsertParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedUser, created)
			})
		})
	})

	t.Run("UpdateUser", func(t *testing.T) {
		t.Parallel()
		existingUser := helpers.GenUser()

		params := user_types.UpsertUserParams{
			Username: "foo",
			Email:    "foo@example.com",
		}

		t.Run("should fail if ValidateUserExists fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			badId := uuid.New()

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, badId).Return(user_types.User{}, user_types.NotFoundError{})

			updated, err := f.underTest.UpdateUser(t.Context(), badId, params)

			assert.Empty(t, updated)
			assert.IsType(t, user_types.NotFoundError{}, err)
		})

		t.Run("should fail if ValidateUser fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(user_types.BadParamsError{})

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, params)

			require.Empty(t, updated)
			assert.IsType(t, user_types.BadParamsError{}, err)
		})

		t.Run("should fail if ValidateUsernameDoesNotConflict fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(user_types.ConflictError{})

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, params)

			assert.Empty(t, updated)
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("should fail if ValidateEmailDoesNotConflict fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(user_types.ConflictError{})
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, params)

			assert.Empty(t, updated)
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("should fail with UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)

			f.userRepoMock.EXPECT().
				UpsertUser(mock.Anything, mock.MatchedBy(helpers.UserMatchesUpsertParams(params))).
				Return(user_types.User{}, errors.New("unknown error"))

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, params)

			assert.Empty(t, updated)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should update a user without optional fields", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			upsertParams := params
			upsertParams.Token = mo.None[string]()
			upsertParams.Bio = mo.None[string]()
			upsertParams.Image = mo.None[string]()

			expectedUpdatedUser := helpers.GenUserWithUpsertParams(upsertParams)
			expectedUpdatedUser.Id = existingUser.Id

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)

			f.userRepoMock.EXPECT().
				UpsertUser(mock.Anything, mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).
				Return(expectedUpdatedUser, nil)

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, upsertParams)

			assert.NoError(t, err)
			assert.Equal(t, expectedUpdatedUser, updated)
		})

		t.Run("should update a user with optional fields", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			upsertParams := params
			upsertParams.Token = mo.Some("updated-token")
			upsertParams.Bio = mo.Some("updated bio")
			upsertParams.Image = mo.Some("http://example.com/updated.png")

			expectedUpdatedUser := helpers.GenUserWithUpsertParams(upsertParams)
			expectedUpdatedUser.Id = existingUser.Id

			f.validationsMock.EXPECT().ValidateUserIdExists(mock.Anything, existingUser.Id).Return(existingUser, nil)
			f.validationsMock.EXPECT().ValidateUser(mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).Return(nil)
			f.validationsMock.EXPECT().ValidateEmailDoesNotConflict(mock.Anything, params.Email).Return(nil)
			f.validationsMock.EXPECT().ValidateUsernameDoesNotConflict(mock.Anything, params.Username).Return(nil)

			f.userRepoMock.EXPECT().
				UpsertUser(mock.Anything, mock.MatchedBy(helpers.UserMatchesUpsertParams(upsertParams))).
				Return(expectedUpdatedUser, nil)

			updated, err := f.underTest.UpdateUser(t.Context(), existingUser.Id, upsertParams)

			assert.NoError(t, err)
			assert.Equal(t, expectedUpdatedUser, updated)
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			id := uuid.New()
			f.userRepoMock.EXPECT().DeleteUser(mock.Anything, id).Return(errors.New("query failure"))

			err := f.underTest.DeleteUser(t.Context(), id)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should delete user successfully", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			id := uuid.New()
			f.userRepoMock.EXPECT().DeleteUser(mock.Anything, id).Return(nil)

			err := f.underTest.DeleteUser(t.Context(), id)
			assert.NoError(t, err)
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError when repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			badEmail := "bad-email@example.com"
			f.userRepoMock.EXPECT().GetUserByEmail(mock.Anything, badEmail).Return(mo.None[user_types.User](), errors.New("query failure"))

			result, err := f.underTest.GetUserByEmail(t.Context(), badEmail)
			assert.Empty(t, result)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return NotFoundError when user does not exist", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			nonExistentEmail := "nonemail@example.com"
			f.userRepoMock.EXPECT().GetUserByEmail(mock.Anything, nonExistentEmail).Return(mo.None[user_types.User](), nil)

			result, err := f.underTest.GetUserByEmail(t.Context(), nonExistentEmail)
			assert.Empty(t, result)
			assert.IsType(t, user_types.NotFoundError{}, err)
		})

		t.Run("should return user when found", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			expectedUser := helpers.GenUser()
			f.userRepoMock.EXPECT().GetUserByEmail(mock.Anything, expectedUser.Email).Return(mo.Some(expectedUser), nil)

			result, err := f.underTest.GetUserByEmail(t.Context(), expectedUser.Email)

			assert.NoError(t, err)
			assert.Equal(t, expectedUser, result)
		})
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError when repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			badUsername := "bad-username"
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, badUsername).Return(mo.None[user_types.User](), errors.New("query failure"))

			result, err := f.underTest.GetUserByUsername(t.Context(), badUsername)
			assert.Empty(t, result)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return NotFoundError when user does not exist", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			nonExistentUsername := "noexisto"
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, nonExistentUsername).Return(mo.None[user_types.User](), nil)

			result, err := f.underTest.GetUserByUsername(t.Context(), nonExistentUsername)
			assert.Empty(t, result)
			assert.IsType(t, user_types.NotFoundError{}, err)
		})

		t.Run("should return user when found", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			expectedUser := helpers.GenUser()
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, expectedUser.Username).Return(mo.Some(expectedUser), nil)

			result, err := f.underTest.GetUserByUsername(t.Context(), expectedUser.Username)

			assert.NoError(t, err)
			assert.Equal(t, expectedUser, result)
		})
	})

	t.Run("IsFollowing", func(t *testing.T) {
		t.Parallel()
		authUser := helpers.GenUser()
		targetUser := helpers.GenUser()

		t.Run("should fail if validate username exists fails", func(t *testing.T) {
			t.Parallel()

			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(user_types.User{}, user_types.NotFoundError{Identifier: "target"})

			res, err := f.underTest.IsFollowing(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.NotFoundError{}, err)
			assert.False(t, res)
		})

		t.Run("should fail if is folllowing fails", func(t *testing.T) {
			t.Parallel()

			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
			f.userRepoMock.EXPECT().IsFollowing(mock.Anything, authUser.Id, targetUser.Id).Return(false, fmt.Errorf("is following failed"))

			res, err := f.underTest.IsFollowing(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.UnknownError{}, err)
			assert.False(t, res)
		})

		t.Run("should return is following successfully", func(t *testing.T) {
			t.Parallel()
			t.Run("when following", func(t *testing.T) {
				t.Parallel()

				// following case
				f := setup(t)
				f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
				f.userRepoMock.EXPECT().IsFollowing(mock.Anything, authUser.Id, targetUser.Id).Return(true, nil)

				res, err := f.underTest.IsFollowing(t.Context(), authUser, targetUser.Username)
				assert.NoError(t, err)
				assert.True(t, res)
			})

			t.Run("when not following", func(t *testing.T) {
				t.Parallel()

				f := setup(t)
				f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
				f.userRepoMock.EXPECT().IsFollowing(mock.Anything, authUser.Id, targetUser.Id).Return(false, nil)

				res, err := f.underTest.IsFollowing(t.Context(), authUser, targetUser.Username)
				assert.NoError(t, err)
				assert.False(t, res)
			})

		})
	})

	t.Run("FollowProfile", func(t *testing.T) {
		t.Parallel()
		authUser := helpers.GenUser()
		targetUser := helpers.GenUser()

		t.Run("should fail if validate profile fails", func(t *testing.T) {
			t.Parallel()

			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(user_types.User{}, user_types.NotFoundError{Identifier: "target"})

			res, err := f.underTest.FollowProfile(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.NotFoundError{}, err)
			assert.Empty(t, res)
		})

		t.Run("should fail if following self", func(t *testing.T) {
			t.Parallel()

			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, authUser.Username).Return(authUser, nil)
			f.validationsMock.EXPECT().ValidateCanFollow(authUser.Id, authUser.Id).Return(user_types.CannotFollowYourselfError{})

			res, err := f.underTest.FollowProfile(t.Context(), authUser, authUser.Username)
			assert.IsType(t, user_types.CannotFollowYourselfError{}, err)
			assert.Empty(t, res)
		})

		t.Run("should fail if follow fails", func(t *testing.T) {
			t.Parallel()

			targetUser := helpers.GenUser()
			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
			f.validationsMock.EXPECT().ValidateCanFollow(authUser.Id, targetUser.Id).Return(nil)

			f.userRepoMock.EXPECT().Follow(mock.Anything, authUser.Id, targetUser.Id).Return(fmt.Errorf("follow failed"))

			res, err := f.underTest.FollowProfile(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.UnknownError{}, err)
			assert.Empty(t, res)
		})

		t.Run("should follow successfully", func(t *testing.T) {
			t.Parallel()

			targetUser := helpers.GenUser()
			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
			f.validationsMock.EXPECT().ValidateCanFollow(authUser.Id, targetUser.Id).Return(nil)
			f.userRepoMock.EXPECT().Follow(mock.Anything, authUser.Id, targetUser.Id).Return(nil)

			res, err := f.underTest.FollowProfile(t.Context(), authUser, targetUser.Username)
			assert.NoError(t, err)
			assert.Equal(t, targetUser, res)
		})
	})

	t.Run("UnfollowProfile", func(t *testing.T) {
		t.Parallel()
		authUser := helpers.GenUser()
		targetUser := helpers.GenUser()

		t.Run("should fail if validate profile fails", func(t *testing.T) {
			t.Parallel()

			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(user_types.User{}, user_types.NotFoundError{Identifier: targetUser.Username})

			res, err := f.underTest.UnfollowProfile(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.NotFoundError{}, err)
			assert.Empty(t, res)
		})

		t.Run("should fail if unfollow fails", func(t *testing.T) {
			t.Parallel()

			targetUser := helpers.GenUser()
			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
			f.userRepoMock.EXPECT().Unfollow(mock.Anything, authUser.Id, targetUser.Id).Return(fmt.Errorf("unfollow failed"))

			res, err := f.underTest.UnfollowProfile(t.Context(), authUser, targetUser.Username)
			assert.IsType(t, user_types.UnknownError{}, err)
			assert.Empty(t, res)
		})

		t.Run("should unfollow successfully", func(t *testing.T) {
			t.Parallel()

			targetUser := helpers.GenUser()
			f := setup(t)
			f.validationsMock.EXPECT().ValidateUsernameExists(mock.Anything, targetUser.Username).Return(targetUser, nil)
			f.userRepoMock.EXPECT().Unfollow(mock.Anything, authUser.Id, targetUser.Id).Return(nil)

			res, err := f.underTest.UnfollowProfile(t.Context(), authUser, targetUser.Username)
			assert.NoError(t, err)
			assert.Equal(t, targetUser, res)
		})
	})
}
