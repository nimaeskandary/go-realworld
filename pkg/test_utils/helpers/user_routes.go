package helpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/stretchr/testify/require"
)

func GetCurrentUserRequest(t *testing.T, authService auth_types.AuthService, authUser user_types.User) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	return WithAuthHeader(t, authService, authUser, req)
}

func UpdateCurrentUserRequest(t *testing.T, authService auth_types.AuthService, authUser user_types.User, updateUser api_gen.UpdateUser) *http.Request {
	body, err := json.Marshal(
		api_gen.UpdateUserRequest{
			User: updateUser,
		})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return WithAuthHeader(t, authService, authUser, req)
}

func CreateUserRequest(t *testing.T, user user_types.User) *http.Request {
	body, err := json.Marshal(
		api_gen.NewUserRequest{
			User: api_gen.NewUser{
				Username: user.Username,
				Email:    user.Email,
				Password: "password",
			},
		})

	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// this is an unauthorized route
	return req
}

func LoginUserRequest(t *testing.T, user user_types.User) *http.Request {
	body, err := json.Marshal(
		api_gen.LoginUserRequest{
			User: api_gen.LoginUser{
				Email:    user.Email,
				Password: "password",
			},
		})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// this is an unauthorized route
	return req
}
