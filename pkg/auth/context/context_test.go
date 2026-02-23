package auth_context_test

import (
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/auth/context"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"

	"github.com/stretchr/testify/assert"
)

func Test_Context(t *testing.T) {
	t.Parallel()

	t.Run("CtxWithUser", func(t *testing.T) {
		t.Parallel()

		t.Run("should set user in context", func(t *testing.T) {
			t.Parallel()

			user := helpers.GenUser()
			ctx := t.Context()
			ctxWithUser := auth_context.CtxWithUser(ctx, user)

			userFromCtx := auth_context.UserFromCtx(ctxWithUser)
			assert.Equal(t, user, userFromCtx.MustGet())
		})
	})

	t.Run("UserFromCtx", func(t *testing.T) {
		t.Parallel()

		t.Run("should return none if user not in context", func(t *testing.T) {
			t.Parallel()

			userFromCtx := auth_context.UserFromCtx(t.Context())
			assert.True(t, userFromCtx.IsNone())
		})
	})
}
