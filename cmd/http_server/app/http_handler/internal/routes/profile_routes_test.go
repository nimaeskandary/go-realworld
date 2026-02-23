package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/fixtures"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ProfileRoutes(t *testing.T) {
	t.Parallel()

	f := fixtures.SetupStandardFixture(t)

	t.Run("GetProfileByUsername", func(t *testing.T) {
		t.Parallel()

		t.Run("should return 422 if profile does not exist", func(t *testing.T) {
			t.Parallel()

			req := helpers.GetProfileByUserNameRequest(t, f.AuthService, mo.None[user_types.User](), "no-existo")
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})

		t.Run("should return profile", func(t *testing.T) {
			t.Parallel()

			users := helpers.CreateUsers(t, f.UserService, 3)
			// have user 0 follow user 1
			_, err := f.UserService.FollowProfile(t.Context(), users[0], users[1].Username)
			require.NoError(t, err)

			type testCase struct {
				name            string
				authUser        mo.Option[user_types.User]
				targetUsername  string
				expectedProfile api_gen.Profile
			}

			testCases := []testCase{
				{
					name:           "when not logged in, should return profile with following=false",
					authUser:       mo.None[user_types.User](),
					targetUsername: users[1].Username,
					expectedProfile: api_gen.Profile{
						Username:  users[1].Username,
						Bio:       users[1].Bio.OrElse(""),
						Image:     users[1].Image.OrElse(""),
						Following: false,
					},
				},
				{
					name:           "when logged in and following the user, should return profile with following=true",
					authUser:       mo.Some(users[0]),
					targetUsername: users[1].Username,
					expectedProfile: api_gen.Profile{
						Username:  users[1].Username,
						Bio:       users[1].Bio.OrElse(""),
						Image:     users[1].Image.OrElse(""),
						Following: true,
					},
				},
				{
					name:           "when logged in and not following the user, should return profile with following=false",
					authUser:       mo.Some(users[2]),
					targetUsername: users[1].Username,
					expectedProfile: api_gen.Profile{
						Username:  users[1].Username,
						Bio:       users[1].Bio.OrElse(""),
						Image:     users[1].Image.OrElse(""),
						Following: false,
					},
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					req := helpers.GetProfileByUserNameRequest(t, f.AuthService, tc.authUser, tc.targetUsername)
					rec := httptest.NewRecorder()
					f.HttpHandler.GetHandler().ServeHTTP(rec, req)

					expected, err := json.Marshal(
						api_gen.GetProfileByUsername200JSONResponse{
							ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{
								Profile: tc.expectedProfile,
							},
						},
					)
					require.NoError(t, err)

					assert.Equal(t, http.StatusOK, rec.Code)
					assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
				})
			}
		})
	})

	t.Run("UnfollowUserByUsername", func(t *testing.T) {
		t.Parallel()

		t.Run("should return 401 if no auth header", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)

			req := helpers.UnfollowUserByUsernameRequest(t, f.AuthService, users[0], "someusername")
			req.Header.Del("Authorization")
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("should return 422 if profile does not exist", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 1)

			req := helpers.UnfollowUserByUsernameRequest(t, f.AuthService, users[0], "no-existo")
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})

		t.Run("should unfollow the user", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)
			var err error
			_, err = f.UserService.FollowProfile(t.Context(), users[0], users[1].Username)
			require.NoError(t, err)

			req := helpers.UnfollowUserByUsernameRequest(t, f.AuthService, users[0], users[1].Username)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			expected, err := json.Marshal(
				api_gen.UnfollowUserByUsername200JSONResponse{
					ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{
						Profile: api_gen.Profile{
							Username:  users[1].Username,
							Bio:       users[1].Bio.OrElse(""),
							Image:     users[1].Image.OrElse(""),
							Following: false,
						},
					},
				},
			)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
		})

		t.Run("should be idempotent if already unfollowing", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)

			req := helpers.UnfollowUserByUsernameRequest(t, f.AuthService, users[0], users[1].Username)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			expected, err := json.Marshal(
				api_gen.UnfollowUserByUsername200JSONResponse{
					ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{
						Profile: api_gen.Profile{
							Username:  users[1].Username,
							Bio:       users[1].Bio.OrElse(""),
							Image:     users[1].Image.OrElse(""),
							Following: false,
						},
					},
				},
			)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
		})
	})

	t.Run("FollowUserByUsername", func(t *testing.T) {
		t.Parallel()

		t.Run("should return 401 if no auth header", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)

			req := helpers.FollowUserByUsernameRequest(t, f.AuthService, users[0], "someusername")
			req.Header.Del("Authorization")
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("should return 422 if profile does not exist", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 1)

			req := helpers.FollowUserByUsernameRequest(t, f.AuthService, users[0], "no-existo")
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})

		t.Run("should return 422 if trying to follow yourself", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 1)

			req := helpers.FollowUserByUsernameRequest(t, f.AuthService, users[0], users[0].Username)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})

		t.Run("should follow the user", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)

			req := helpers.FollowUserByUsernameRequest(t, f.AuthService, users[0], users[1].Username)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			expected, err := json.Marshal(
				api_gen.FollowUserByUsername200JSONResponse{
					ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{
						Profile: api_gen.Profile{
							Username:  users[1].Username,
							Bio:       users[1].Bio.OrElse(""),
							Image:     users[1].Image.OrElse(""),
							Following: true,
						},
					},
				},
			)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
		})

		t.Run("should be idempotent if already following", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)
			var err error
			_, err = f.UserService.FollowProfile(t.Context(), users[0], users[1].Username)
			require.NoError(t, err)

			req := helpers.FollowUserByUsernameRequest(t, f.AuthService, users[0], users[1].Username)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			expected, err := json.Marshal(
				api_gen.FollowUserByUsername200JSONResponse{
					ProfileResponseJSONResponse: api_gen.ProfileResponseJSONResponse{
						Profile: api_gen.Profile{
							Username:  users[1].Username,
							Bio:       users[1].Bio.OrElse(""),
							Image:     users[1].Image.OrElse(""),
							Following: true,
						},
					},
				},
			)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
		})
	})
}
