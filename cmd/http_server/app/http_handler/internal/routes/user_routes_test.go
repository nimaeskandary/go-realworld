package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/http_handler/internal/transformers"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/fixtures"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UserRoutes(t *testing.T) {
	t.Parallel()

	f := fixtures.SetupStandardFixture(t)

	t.Run("GetCurrentUser", func(t *testing.T) {
		t.Parallel()
		user := helpers.CreateUsers(t, f.UserService, 1)[0]

		t.Run("should return a 401 if there is no auth header", func(t *testing.T) {
			t.Parallel()

			req := helpers.GetCurrentUserRequest(t, f.AuthService, user)
			req.Header.Del("Authorization")

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("should return a 200 on success", func(t *testing.T) {
			t.Parallel()

			req := helpers.GetCurrentUserRequest(t, f.AuthService, user)

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			validateUserResponse(t, f.AuthService, rec.Body.Bytes(), user)
		})
	})

	t.Run("UpdateCurrentUser", func(t *testing.T) {
		t.Parallel()

		t.Run("should return a 401 if there is no auth header", func(t *testing.T) {
			t.Parallel()
			user := helpers.CreateUsers(t, f.UserService, 1)[0]

			updatedUsername := "updatedusername"
			req := helpers.UpdateCurrentUserRequest(t, f.AuthService, user, api_gen.UpdateUser{Username: &updatedUsername})
			req.Header.Del("Authorization")

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("should return a 422 if there was a conflict", func(t *testing.T) {
			t.Parallel()
			users := helpers.CreateUsers(t, f.UserService, 2)

			req := helpers.UpdateCurrentUserRequest(t, f.AuthService, users[0], api_gen.UpdateUser{Username: &users[1].Username})

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, string(rec.Body.String()), "username already exists")
		})

		t.Run("should return a 422 if bad params", func(t *testing.T) {
			t.Parallel()
			user := helpers.CreateUsers(t, f.UserService, 1)[0]

			updatedEmail := "invalidemail"
			req := helpers.UpdateCurrentUserRequest(t, f.AuthService, user, api_gen.UpdateUser{Email: &updatedEmail})

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, string(rec.Body.String()), "Email")
		})

		t.Run("should successfully update partial fields", func(t *testing.T) {
			t.Parallel()
			user := helpers.CreateUsers(t, f.UserService, 1)[0]

			expectedUpdatedUser := user
			expectedUpdatedUser.Username = "updatedusername"
			req := helpers.UpdateCurrentUserRequest(t, f.AuthService, user, api_gen.UpdateUser{Username: &expectedUpdatedUser.Username})

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			validateUserResponse(t, f.AuthService, rec.Body.Bytes(), expectedUpdatedUser)
		})

		t.Run("should successfully update all fields", func(t *testing.T) {
			t.Parallel()
			user := helpers.CreateUsers(t, f.UserService, 1)[0]

			expectedUpdatedUser := user
			expectedUpdatedUser.Username = "otherupdatedusername"
			expectedUpdatedUser.Email = "updatedemail@example.com"
			bio := "updated-bio"
			expectedUpdatedUser.Bio = mo.Some(bio)
			image := "https://example.com/updated-image.jpg"
			expectedUpdatedUser.Image = mo.Some(image)

			req := helpers.UpdateCurrentUserRequest(t, f.AuthService, user, api_gen.UpdateUser{
				Username: &expectedUpdatedUser.Username,
				Email:    &expectedUpdatedUser.Email,
				Bio:      &bio,
				Image:    &image,
			})

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			validateUserResponse(t, f.AuthService, rec.Body.Bytes(), expectedUpdatedUser)
		})
	})

	t.Run("CreateUser", func(t *testing.T) {
		t.Parallel()

		type createUserFixture struct {
			fixtures.StandardFixture
			user user_types.User
		}

		var setup = func() createUserFixture {
			user := helpers.GenUser()
			return createUserFixture{
				StandardFixture: f,
				user:            user,
			}
		}

		t.Run("should return a 422 if there is a conflict", func(t *testing.T) {
			t.Parallel()
			f := setup()

			_, err := f.UserService.CreateUser(t.Context(), user_types.UpsertUserParams{
				Username: f.user.Username,
				Email:    f.user.Email,
			})
			require.NoError(t, err)
			t.Cleanup(func() { _ = f.UserService.DeleteUser(context.Background(), f.user.Id) })

			req := helpers.CreateUserRequest(t, f.user)
			// ensure we are testing that this is an unauthorized route
			req.Header.Del("Authorization")

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, string(rec.Body.String()), "already exists")
		})

		t.Run("should return a 422 if there are bad params", func(t *testing.T) {
			t.Parallel()
			f := setup()

			invalidEmail := "invalidemail"
			f.user.Email = invalidEmail

			req := helpers.CreateUserRequest(t, f.user)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, string(rec.Body.String()), "Email")
		})

		t.Run("should successfully create a user", func(t *testing.T) {
			t.Parallel()
			f := setup()

			req := helpers.CreateUserRequest(t, f.user)
			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)

			result := new(api_gen.CreateUser201JSONResponse)
			err := json.Unmarshal(rec.Body.Bytes(), result)
			require.NoError(t, err)

			assert.Equal(t, f.user.Username, result.User.Username)
			assert.Equal(t, f.user.Email, result.User.Email)

			expectedUser, err := f.UserService.GetUserByUsername(t.Context(), result.User.Username)
			require.NoError(t, err)
			validateUserResponse(t, f.AuthService, rec.Body.Bytes(), expectedUser)
		})
	})

	// the login route for this project just tests that the user exists, no password validation
	t.Run("Login", func(t *testing.T) {
		t.Parallel()

		user, err := f.UserService.CreateUser(t.Context(), user_types.UpsertUserParams{
			Username: "loginuser",
			Email:    "loginuser@example.com",
		})
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.UserService.DeleteUser(context.Background(), user.Id) })

		t.Run("should return a 422 if the user is not found", func(t *testing.T) {
			t.Parallel()
			user := helpers.GenUser()
			user.Username = "nonexistentuser"
			user.Email = "nonexistentuser@example.com"

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, helpers.LoginUserRequest(t, user))

			expected, err := json.Marshal(
				api_gen.Login422JSONResponse{
					GenericErrorJSONResponse: transformers.ToApiError(fmt.Errorf("user not found")),
				},
			)
			require.NoError(t, err)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, expected, bytes.TrimSpace(rec.Body.Bytes()))
		})

		t.Run("should return a 200 on success", func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			f.HttpHandler.GetHandler().ServeHTTP(rec, helpers.LoginUserRequest(t, user))

			assert.Equal(t, http.StatusOK, rec.Code)
			validateUserResponse(t, f.AuthService, rec.Body.Bytes(), user)
		})
	})
}

func validateUserResponse(t *testing.T, authService auth_types.AuthService, responseBody []byte, expectedUser user_types.User) {
	actual := new(api_gen.UserResponseJSONResponse)
	err := json.Unmarshal(responseBody, actual)
	require.NoError(t, err)

	assert.Equal(t, expectedUser.Bio.OrElse(""), actual.User.Bio)
	assert.Equal(t, expectedUser.Email, actual.User.Email)
	assert.Equal(t, expectedUser.Image.OrElse(""), actual.User.Image)
	assert.Equal(t, expectedUser.Username, actual.User.Username)

	token, err := authService.ParseToken(t.Context(), actual.User.Token)
	require.NoError(t, err)
	assert.Equal(t, expectedUser.Id, token.GetUser().Id)
}
