package internal_test

import (
	"errors"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"
	"github.com/nimaeskandary/go-realworld/pkg/user/internal"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"
	user_types_mocks "github.com/nimaeskandary/go-realworld/pkg/user/types/mocks"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UserValidationsImpl(t *testing.T) {
	t.Parallel()

	type testFixture struct {
		userRepoMock *user_types_mocks.MockUserRepository
		underTest    user_types.UserValidations
	}

	setup := func(t *testing.T) testFixture {
		userRepoMock := user_types_mocks.NewMockUserRepository(t)
		underTest := internal.NewUserValidationsImpl(userRepoMock)
		return testFixture{
			userRepoMock: userRepoMock,
			underTest:    underTest,
		}
	}

	user := helpers.GenUser()

	t.Run("ValidateUser", func(t *testing.T) {
		t.Parallel()

		t.Run("should return no error if user is valid", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			err := f.underTest.ValidateUser(user)
			assert.NoError(t, err)
		})

		t.Run("should return BadParamsError if username is invalid", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			badUser := user
			badUser.Username = ""

			err := f.underTest.ValidateUser(badUser)
			assert.IsType(t, user_types.BadParamsError{}, err)
		})

		t.Run("should return BadParamsError if email is invalid", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			badUser := user
			badUser.Email = ""

			err := f.underTest.ValidateUser(badUser)
			assert.IsType(t, user_types.BadParamsError{}, err)

			badUser.Email = "invalid-email"

			err = f.underTest.ValidateUser(badUser)
			assert.IsType(t, user_types.BadParamsError{}, err)
		})
	})

	t.Run("ValidateUsernameDoesNotConflict", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByUsername(mock.Anything, "error-case").
				Return(mo.None[user_types.User](), errors.New("db down"))

			err := f.underTest.ValidateUsernameDoesNotConflict(t.Context(), "error-case")
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return ConflictError if username already exists", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByUsername(mock.Anything, "exists").
				Return(mo.Some(helpers.GenUser()), nil)

			err := f.underTest.ValidateUsernameDoesNotConflict(t.Context(), "exists")
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("should return no error if username is free", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByUsername(mock.Anything, "new-user").
				Return(mo.None[user_types.User](), nil)

			err := f.underTest.ValidateUsernameDoesNotConflict(t.Context(), "new-user")
			assert.NoError(t, err)
		})
	})

	t.Run("ValidateEmailDoesNotConflict", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByEmail(mock.Anything, "error-case@test.com").
				Return(mo.None[user_types.User](), errors.New("db down"))

			err := f.underTest.ValidateEmailDoesNotConflict(t.Context(), "error-case@test.com")
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return ConflictError if email already exists", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByEmail(mock.Anything, "exists@test.com").
				Return(mo.Some(helpers.GenUser()), nil)

			err := f.underTest.ValidateEmailDoesNotConflict(t.Context(), "exists@test.com")
			assert.IsType(t, user_types.ConflictError{}, err)
		})

		t.Run("should return no error if email is free", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().
				GetUserByEmail(mock.Anything, "new-email@test.com").
				Return(mo.None[user_types.User](), nil)

			err := f.underTest.ValidateEmailDoesNotConflict(t.Context(), "new-email@test.com")
			assert.NoError(t, err)
		})
	})

	t.Run("ValidateUserIdExists", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			id := uuid.New()
			f.userRepoMock.EXPECT().GetUserById(mock.Anything, id).Return(mo.None[user_types.User](), errors.New("db down"))

			user, err := f.underTest.ValidateUserIdExists(t.Context(), id)
			assert.Empty(t, user)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return NotFoundError if user does not exist", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			id := uuid.New()
			f.userRepoMock.EXPECT().GetUserById(mock.Anything, id).Return(mo.None[user_types.User](), nil)

			user, err := f.underTest.ValidateUserIdExists(t.Context(), id)
			assert.Empty(t, user)
			assert.IsType(t, user_types.NotFoundError{}, err)
		})

		t.Run("should return user if found", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().GetUserById(mock.Anything, user.Id).Return(mo.Some(user), nil)

			user, err := f.underTest.ValidateUserIdExists(t.Context(), user.Id)
			assert.NoError(t, err)
			assert.Equal(t, user, user)
		})
	})

	t.Run("ValidateUsernameExists", func(t *testing.T) {
		t.Parallel()

		t.Run("should return UnknownError if repo fails", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			username := "test-username"
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, username).Return(mo.None[user_types.User](), errors.New("db down"))

			user, err := f.underTest.ValidateUsernameExists(t.Context(), username)
			assert.Empty(t, user)
			assert.IsType(t, user_types.UnknownError{}, err)
		})

		t.Run("should return NotFoundError if user does not exist", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			username := "test-username"
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, username).Return(mo.None[user_types.User](), nil)

			user, err := f.underTest.ValidateUsernameExists(t.Context(), username)
			assert.Empty(t, user)
			assert.IsType(t, user_types.NotFoundError{}, err)
		})

		t.Run("should return user if found", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			f.userRepoMock.EXPECT().GetUserByUsername(mock.Anything, user.Username).Return(mo.Some(user), nil)

			user, err := f.underTest.ValidateUsernameExists(t.Context(), user.Username)
			assert.NoError(t, err)
			assert.Equal(t, user, user)
		})
	})

	t.Run("ValidateCanFollow", func(t *testing.T) {
		t.Parallel()

		t.Run("should return CannotFollowYourselfError if user ids are equal", func(t *testing.T) {
			t.Parallel()
			f := setup(t)
			id := uuid.New()

			err := f.underTest.ValidateCanFollow(id, id)
			assert.IsType(t, user_types.CannotFollowYourselfError{}, err)
		})

		t.Run("should return no error if user ids are different", func(t *testing.T) {
			t.Parallel()
			f := setup(t)

			err := f.underTest.ValidateCanFollow(uuid.New(), uuid.New())
			assert.Empty(t, err)
		})

	})
}
