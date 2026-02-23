package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimaeskandary/go-realworld/cmd/http_server/app/middleware"
	obs_types_mocks "github.com/nimaeskandary/go-realworld/pkg/observability/types/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_WithCorsMiddleware(t *testing.T) {
	t.Parallel()

	allowedOrigins := []string{"http://example.com", "http://another.com"}
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware.WithCorsMiddleware(allowedOrigins, nextHandler)

	t.Run("should allow requests from allowed origins and set CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
		assert.True(t, nextCalled)
	})

	t.Run("should handle preflight OPTIONS requests and return 204", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.False(t, nextCalled)
	})

	t.Run("should block requests from disallowed origins with 403", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://notallowed.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, nextCalled)
	})
}

func Test_HandleError(t *testing.T) {
	t.Parallel()

	type fixture struct {
		logger *obs_types_mocks.MockLogger
	}

	var setupFixture = func(t *testing.T) fixture {
		return fixture{
			logger: obs_types_mocks.NewMockLogger(t),
		}
	}

	t.Run("should pass through response if no error", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			return "success", nil
		}

		underTest := middleware.HandleError(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)

		rec := httptest.NewRecorder()
		resp, err := handler(t.Context(), rec, httptest.NewRequest(http.MethodGet, "/user", nil), nil)

		assert.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled)
	})

	t.Run("should log unexpected errors and return generic error in its place", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		expectedError := fmt.Errorf("big issue!")

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			return nil, expectedError
		}

		f.logger.EXPECT().Error(
			mock.Anything,
			"unexpected error handling request",
			expectedError,
			[]any{"operation_id", string(middleware.GetCurrentUserOpId)},
		)

		underTest := middleware.HandleError(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)

		rec := httptest.NewRecorder()
		resp, err := handler(t.Context(), rec, httptest.NewRequest(http.MethodGet, "/user", nil), nil)

		assert.True(t, nextCalled)
		assert.EqualError(t, err, ("unexpected error occured"))
		assert.Empty(t, resp)
	})

	t.Run("should recover from panics and return generic error in its place", func(t *testing.T) {
		t.Parallel()
		f := setupFixture(t)

		expectedError := fmt.Errorf("big issue!")

		nextCalled := false
		nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
			nextCalled = true
			panic(expectedError)
		}

		f.logger.EXPECT().Error(
			mock.Anything,
			"panic recovered in middleware",
			fmt.Errorf("panic: %v", expectedError),
			[]any{"operation_id", string(middleware.GetCurrentUserOpId)},
		)

		underTest := middleware.HandleError(f.logger)

		handler := underTest(
			nextFn,
			string(middleware.GetCurrentUserOpId),
		)

		rec := httptest.NewRecorder()
		resp, err := handler(t.Context(), rec, httptest.NewRequest(http.MethodGet, "/user", nil), nil)

		assert.True(t, nextCalled)
		assert.EqualError(t, err, ("unexpected error occured"))
		assert.Empty(t, resp)
	})
}
