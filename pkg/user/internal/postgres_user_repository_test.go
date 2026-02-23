package internal_test

import (
	"testing"
	"time"

	"github.com/nimaeskandary/go-realworld/pkg/test_utils/fixtures"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

func Test_PostgressUserRepository(t *testing.T) {
	t.Parallel()

	f := fixtures.SetupStandardFixture(t)
	underTest := f.UserRepo

	t.Run("UpsertUser", func(t *testing.T) {
		t.Parallel()
		t.Run("should insert new user with all fields", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			user.Bio = mo.Some("test bio")
			user.Image = mo.Some("http://example.com/image.png")

			created, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)
			assert.Equal(t, user, created)

			fromDb, err := underTest.GetUserById(t.Context(), user.Id)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsSome())
			assert.Equal(t, user, fromDb.MustGet())
		})

		t.Run("should insert a user without optional fields", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			user.Bio = mo.None[string]()
			user.Image = mo.None[string]()

			created, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)
			assert.Equal(t, user, created)
		})

		t.Run("should update existing user with optional fields", func(t *testing.T) {
			t.Parallel()

			created, err := underTest.UpsertUser(t.Context(), helpers.GenUser())
			assert.NoError(t, err)

			updatedUser := user_types.User{
				Id:              created.Id,
				Username:        "updateduseroptional",
				Email:           "updatedoptional@example.com",
				Bio:             mo.Some("updated bio"),
				Image:           mo.Some("http://example.com/updated-image.png"),
				CreatedAtMillis: created.CreatedAtMillis,
				UpdatedAtMillis: time.Now().UnixMilli(),
			}

			updated, err := underTest.UpsertUser(t.Context(), updatedUser)
			assert.NoError(t, err)
			assert.Equal(t, updatedUser, updated)
		})

		t.Run("should update existing user without optional fields", func(t *testing.T) {
			t.Parallel()

			created, err := underTest.UpsertUser(t.Context(), helpers.GenUser())
			assert.NoError(t, err)

			updatedUser := user_types.User{
				Id:              created.Id,
				Username:        "updateduser",
				Email:           "updated@example.com",
				Bio:             mo.None[string](),
				Image:           mo.None[string](),
				CreatedAtMillis: created.CreatedAtMillis,
				UpdatedAtMillis: time.Now().UnixMilli(),
			}

			updated, err := underTest.UpsertUser(t.Context(), updatedUser)
			assert.NoError(t, err)
			assert.Equal(t, updatedUser, updated)
		})
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Parallel()

		t.Run("should return no error when deleting non-existing user", func(t *testing.T) {
			t.Parallel()

			err := underTest.DeleteUser(t.Context(), uuid.New())
			assert.NoError(t, err)
		})

		t.Run("should delete existing user", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			_, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)

			err = underTest.DeleteUser(t.Context(), user.Id)
			assert.NoError(t, err)

			fromDb, err := underTest.GetUserById(t.Context(), user.Id)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsNone())
		})
	})

	t.Run("GetUserById", func(t *testing.T) {
		t.Parallel()

		t.Run("should return none for non-existing user", func(t *testing.T) {
			t.Parallel()

			result, err := underTest.GetUserById(t.Context(), uuid.New())
			assert.NoError(t, err)
			assert.True(t, result.IsNone())
		})

		t.Run("should return existing user", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			_, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)

			fromDb, err := underTest.GetUserById(t.Context(), user.Id)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsSome())
			assert.Equal(t, user, fromDb.MustGet())
		})
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		t.Parallel()

		t.Run("should return none for non-existing user", func(t *testing.T) {
			t.Parallel()

			result, err := underTest.GetUserByUsername(t.Context(), "nonexistinguser")
			assert.NoError(t, err)
			assert.True(t, result.IsNone())
		})

		t.Run("should return existing user", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			created, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)

			fromDb, err := underTest.GetUserByUsername(t.Context(), created.Username)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsSome())
			assert.Equal(t, user, fromDb.MustGet())
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		t.Parallel()

		t.Run("should return none for non-existing user", func(t *testing.T) {
			t.Parallel()

			result, err := underTest.GetUserByEmail(t.Context(), "nonexisting@example.com")
			assert.NoError(t, err)
			assert.True(t, result.IsNone())
		})

		t.Run("should return existing user", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			_, err := underTest.UpsertUser(t.Context(), user)
			assert.NoError(t, err)

			fromDb, err := underTest.GetUserByEmail(t.Context(), user.Email)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsSome())
			assert.Equal(t, user, fromDb.MustGet())
		})
	})

	t.Run("IsFollowing", func(t *testing.T) {
		t.Parallel()

		users := helpers.CreateUsers(t, f.UserService, 2)

		t.Run("should return false when not following", func(t *testing.T) {
			t.Parallel()

			isFollowing, err := f.UserRepo.IsFollowing(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)
			assert.False(t, isFollowing)
		})

		t.Run("should return true when following", func(t *testing.T) {
			t.Parallel()

			err := f.UserRepo.Follow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			isFollowing, err := f.UserRepo.IsFollowing(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)
			assert.True(t, isFollowing)
		})
	})

	t.Run("Follow", func(t *testing.T) {
		t.Parallel()

		users := helpers.CreateUsers(t, f.UserService, 2)

		t.Run("should be idempotent if already following", func(t *testing.T) {
			t.Parallel()

			err := f.UserRepo.Follow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			err = f.UserRepo.Follow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			isFollowing, err := f.UserRepo.IsFollowing(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)
			assert.True(t, isFollowing)
		})
	})

	t.Run("Unfollow", func(t *testing.T) {
		t.Parallel()

		users := helpers.CreateUsers(t, f.UserService, 2)

		t.Run("should unfollow successfully", func(t *testing.T) {
			t.Parallel()

			err := f.UserRepo.Follow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			err = f.UserRepo.Unfollow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			isFollowing, err := f.UserRepo.IsFollowing(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)
			assert.False(t, isFollowing)
		})

		t.Run("should be idempotent if already not followed", func(t *testing.T) {
			t.Parallel()

			err := f.UserRepo.Unfollow(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)

			isFollowing, err := f.UserRepo.IsFollowing(t.Context(), users[0].Id, users[1].Id)
			assert.NoError(t, err)
			assert.False(t, isFollowing)
		})
	})
}
