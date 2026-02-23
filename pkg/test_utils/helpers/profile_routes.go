package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func GetProfileByUserNameRequest(
	t *testing.T,
	authService auth_types.AuthService,
	authUser mo.Option[user_types.User],
	targetUsername string) *http.Request {
	body, err := json.Marshal(
		api_gen.GetProfileByUsernameRequestObject{
			Username: targetUsername,
		})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/profiles/%v", targetUsername), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if authUser.IsSome() {
		return WithAuthHeader(t, authService, authUser.MustGet(), req)
	}

	return req
}

func UnfollowUserByUsernameRequest(
	t *testing.T,
	authService auth_types.AuthService,
	authUser user_types.User,
	targetUsername string) *http.Request {

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/profiles/%v/follow", targetUsername), nil)
	return WithAuthHeader(t, authService, authUser, req)
}

func FollowUserByUsernameRequest(
	t *testing.T,
	authService auth_types.AuthService,
	authUser user_types.User,
	targetUsername string) *http.Request {

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/profiles/%v/follow", targetUsername), nil)
	return WithAuthHeader(t, authService, authUser, req)
}
