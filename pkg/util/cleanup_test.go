package util_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nimaeskandary/go-realworld/pkg/util"
	"github.com/stretchr/testify/assert"
)

func Test_CleanupManager(t *testing.T) {
	t.Parallel()

	t.Run("should execute registered functions in reverse order", func(t *testing.T) {
		t.Parallel()
		underTest := util.NewCleanupManager(t.Context(), false)

		var executionOrder []int
		underTest.RegisterCleanupFunc(func() { executionOrder = append(executionOrder, 1) })
		underTest.RegisterCleanupFunc(func() { executionOrder = append(executionOrder, 2) })
		underTest.RegisterCleanupFunc(func() { executionOrder = append(executionOrder, 3) })

		underTest.Cleanup()

		assert.Equal(t, []int{3, 2, 1}, executionOrder)
	})

	t.Run("should only execute cleanup functions once (idempotency)", func(t *testing.T) {
		t.Parallel()
		underTest := util.NewCleanupManager(t.Context(), false)

		callCount := atomic.Int32{}
		underTest.RegisterCleanupFunc(func() {
			callCount.Add(1)
		})

		underTest.Cleanup()
		underTest.Cleanup()

		assert.Equal(t, int32(1), callCount.Load())
	})

	t.Run("should cleanup when context is cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		underTest := util.NewCleanupManager(ctx, true)

		callCount := atomic.Int32{}
		cleanedUp := make(chan struct{})
		underTest.RegisterCleanupFunc(func() {
			callCount.Add(1)
			close(cleanedUp)
		})

		cancel()
		underTest.Cleanup()

		select {
		case <-cleanedUp:
			assert.Equal(t, int32(1), callCount.Load())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("cleanup was not called")
		}
	})
}
