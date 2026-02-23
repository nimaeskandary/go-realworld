package helpers

import (
	"fmt"
	"net/http"
	"testing"

	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	user_types "github.com/nimaeskandary/go-realworld/pkg/user/types"

	"github.com/stretchr/testify/require"
)

func WithAuthHeader(t *testing.T, authService auth_types.AuthService, authUser user_types.User, req *http.Request) *http.Request {
	token, err := authService.GenerateToken(t.Context(), authUser.Username)
	require.NoError(t, err)

	req.Header.Set("Authorization", fmt.Sprintf("Token %v", token.GetTokenString()))
	return req
}
