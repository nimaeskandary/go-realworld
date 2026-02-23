package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/middleware"
	"github.com/nimaeskandary/go-realworld/pkg/api_gen"
	"github.com/nimaeskandary/go-realworld/pkg/auth/context"
	"github.com/nimaeskandary/go-realworld/pkg/auth/types"
	"github.com/nimaeskandary/go-realworld/pkg/auth/types/mocks"
	"github.com/nimaeskandary/go-realworld/pkg/observability/types/mocks"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateAuthContext(t *testing.T) {
	t.Parallel()

	type fixture struct {
		authService *auth_types_mocks.MockAuthService
		logger      *obs_types_mocks.MockLogger
	}

	var setupFixture = func(t *testing.T) fixture {
		return fixture{
			authService: auth_types_mocks.NewMockAuthService(t),
			logger:      obs_types_mocks.NewMockLogger(t),
		}
	}

	t.Run("should just return next if there is no auth header", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			// we expect no user to be set in context
			userFromContext := auth_context.UserFromCtx(ctx)
			assert.True(t, userFromContext.IsNone())
			return "success", nil
		}

		underTest := middleware.CreateAuthContext(f.logger, f.authService)

		handler := underTest(
			nextFn,
			string(middleware.LoginOpId),
		)
		rec := httptest.NewRecorder()
		req := helpers.LoginUserRequest(t, helpers.GenUser())
		assert.Empty(t, req.Header.Get("Authorization"))

		resp, err := handler(
			t.Context(),
			rec,
			req,
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled)
	})

	t.Run("should return unauthorized if auth token is invalid", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name         string
			authHeader   string
			parsingError auth_types.DomainError
		}

		testCases := []testCase{
			{
				name:         "invalid token prefix",
				authHeader:   "some invalid",
				parsingError: auth_types.InvalidTokenError{},
			},
			{
				name:         "invalid token",
				authHeader:   "Token invalid",
				parsingError: auth_types.InvalidTokenError{},
			},
			{
				name:         "expired token",
				authHeader:   "Token expired",
				parsingError: auth_types.ExpiredTokenError{},
			},
			{
				name:         "unknown error",
				authHeader:   "Token error-unknown",
				parsingError: auth_types.UnknownError{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				f := setupFixture(t)

				nextCalled := false
				nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
					nextCalled = true
					return "success", nil
				}

				expectedToken := strings.TrimPrefix(tc.authHeader, "Token ")
				if tc.parsingError != nil {
					f.authService.EXPECT().ParseToken(mock.Anything, expectedToken).Return(nil, tc.parsingError)
				}
				f.logger.EXPECT().Error(mock.Anything, "error parsing auth token", tc.parsingError, []any{"token", expectedToken})

				underTest := middleware.CreateAuthContext(f.logger, f.authService)

				handler := underTest(
					nextFn,
					string(middleware.GetCurrentUserOpId),
				)
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/user", nil)
				req.Header.Set("Authorization", tc.authHeader)
				resp, err := handler(
					t.Context(),
					rec,
					req,
					nil,
				)

				assert.NoError(t, err)
				assert.Equal(t, api_gen.UnauthorizedResponse{}, resp)
				assert.False(t, nextCalled)
			})
		}
	})

	t.Run("should set user in context if valid auth header is present", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		expectedUser := helpers.GenUser()
		expectedParsedToken := auth_types_mocks.NewMockAuthToken(t)

		f.authService.EXPECT().ParseToken(mock.Anything, "valid").Return(expectedParsedToken, nil)
		f.logger.EXPECT().CtxWithLogAttributes(mock.Anything, []any{"user_id", expectedUser.Id.String(), "user_username", expectedUser.Username}).Return(t.Context())
		// since we will inject user into context
		expectedParsedToken.EXPECT().GetUser().Return(expectedUser)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			// expect to find the user in the the context
			userFromContext := auth_context.UserFromCtx(ctx)
			assert.Equal(t, expectedUser, userFromContext.MustGet())
			return "success", nil
		}

		underTest := middleware.CreateAuthContext(f.logger, f.authService)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		req.Header.Set("Authorization", "Token valid")
		resp, err := handler(
			t.Context(),
			rec,
			req,
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled)
	})
}

func Test_AuthenticateRoute(t *testing.T) {
	t.Parallel()

	type fixture struct {
		logger *obs_types_mocks.MockLogger
	}

	var setupFixture = func(t *testing.T) fixture {
		return fixture{
			logger: obs_types_mocks.NewMockLogger(t),
		}
	}

	t.Run("should just return next if the operation is non authenticated", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			return "success", nil
		}

		underTest := middleware.AuthenticateRoute(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.LoginOpId),
		)
		rec := httptest.NewRecorder()
		resp, err := handler(
			t.Context(),
			rec,
			helpers.LoginUserRequest(t, helpers.GenUser()),
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled)
	})

	t.Run("should return unauthorized if no auth user in context", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			return "success", nil
		}

		underTest := middleware.AuthenticateRoute(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		resp, err := handler(
			t.Context(),
			rec,
			req,
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, api_gen.UnauthorizedResponse{}, resp)
		assert.False(t, nextCalled)
	})

	t.Run("should call next for authenticated route when auth user in context", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			return "success", nil
		}

		underTest := middleware.AuthenticateRoute(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		resp, err := handler(
			auth_context.CtxWithUser(t.Context(), helpers.GenUser()),
			rec,
			req,
			nil,
		)

		assert.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled)
	})
}
