package util_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nimaeskandary/go-realworld/pkg/util"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

type TestInterface interface {
	util.FxLifecycle
	DoSomething() string
}

type testImpl struct {
	startCalled bool
	stopCalled  bool
	failStart   bool
	failStop    bool
}

func NewTestImpl(failStart bool, failStop bool) *testImpl {
	return &testImpl{
		failStart: failStart,
		failStop:  failStop,
	}
}

func (t *testImpl) DoSomething() string { return "done" }

func (l *testImpl) Start(ctx context.Context) error {
	l.startCalled = true
	if l.failStart {
		return errors.New("start failed")
	}
	return nil
}

func (l *testImpl) Stop(ctx context.Context) error {
	l.stopCalled = true
	if l.failStop {
		return errors.New("stop failed")
	}
	return nil
}

func Test_DI(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		t.Run("should provide a component as an interface", func(t *testing.T) {
			t.Parallel()

			var extracted TestInterface
			app := util.CreateFxAppAndExtract(
				[]fx.Option{util.NewFxModule[TestInterface]("test-module", func() *testImpl {
					return NewTestImpl(false, false)
				})},
				&extracted,
			)

			assert.NoError(t, app.Start(t.Context()))

			assert.NotNil(t, extracted)
			assert.Equal(t, "done", extracted.DoSomething())

			assert.NoError(t, app.Stop(t.Context()))
		})

		t.Run("should respect passed fx.Options", func(t *testing.T) {
			t.Parallel()

			type dep string

			var extracted TestInterface
			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					util.NewFxModule[TestInterface](
						"opt-test",
						// this would fail if dep wasn't properly injected
						func(input dep) *testImpl {
							return NewTestImpl(false, false)
						},
						fx.Provide(func() dep { return "foo" }),
					),
				},
				&extracted,
			)

			assert.NoError(t, app.Start(t.Context()))

			assert.NotNil(t, extracted)
			assert.Equal(t, "done", extracted.DoSomething())

			assert.NoError(t, app.Stop(t.Context()))

		})
	})

	t.Run("NewWithLifeCycle", func(t *testing.T) {
		t.Parallel()

		t.Run("should provide a component as an interface", func(t *testing.T) {
			t.Parallel()

			var extracted TestInterface
			app := util.CreateFxAppAndExtract(
				[]fx.Option{util.NewFxModuleWithLifecycle[TestInterface]("test-module", func() *testImpl {
					return NewTestImpl(false, false)
				})},
				&extracted,
			)

			assert.NoError(t, app.Start(t.Context()))

			assert.NotNil(t, extracted)
			assert.Equal(t, "done", extracted.DoSomething())

			assert.NoError(t, app.Stop(t.Context()))
		})

		t.Run("should respect passed fx.Options", func(t *testing.T) {
			t.Parallel()

			type dep string

			var extracted TestInterface
			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					util.NewFxModuleWithLifecycle[TestInterface](
						"opt-test",
						// this would fail if dep wasn't properly injected
						func(input dep) *testImpl {
							return NewTestImpl(false, false)
						},
						fx.Provide(func() dep { return "foo" }),
					),
				},
				&extracted,
			)

			assert.NoError(t, app.Start(t.Context()))

			assert.NotNil(t, extracted)
			assert.Equal(t, "done", extracted.DoSomething())

			assert.NoError(t, app.Stop(t.Context()))
		})

		t.Run("should trigger Start and Stop hooks", func(t *testing.T) {
			t.Parallel()

			var extracted TestInterface
			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					util.NewFxModuleWithLifecycle[TestInterface]("lifecycle-module", func() *testImpl {
						return NewTestImpl(false, false)
					}),
				},
				&extracted,
			)

			assert.NoError(t, app.Start(t.Context()))
			assert.True(t, extracted.(*testImpl).startCalled)

			assert.NoError(t, app.Stop(t.Context()))
			assert.True(t, extracted.(*testImpl).stopCalled)
		})

		t.Run("should propagate start errors", func(t *testing.T) {
			t.Parallel()

			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					util.NewFxModuleWithLifecycle[TestInterface]("fail-module", func() *testImpl {
						return NewTestImpl(true, false)
					}),
				},
			)

			err := app.Start(t.Context())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "start failed")
		})

		t.Run("should propagate stop errors", func(t *testing.T) {
			t.Parallel()

			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					util.NewFxModuleWithLifecycle[TestInterface]("fail-module", func() *testImpl {
						return NewTestImpl(false, true)
					}),
				},
			)

			assert.NoError(t, app.Start(t.Context()))

			err := app.Stop(t.Context())
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "stop failed")
		})
	})

	t.Run("CreateFxAppAndExtract", func(t *testing.T) {
		t.Parallel()

		t.Run("should extract multiple dependencies", func(t *testing.T) {
			t.Parallel()

			type DepA struct{}
			type DepB struct{}

			var a *DepA
			var b *DepB

			app := util.CreateFxAppAndExtract(
				[]fx.Option{
					fx.Provide(func() *DepA { return &DepA{} }),
					fx.Provide(func() *DepB { return &DepB{} }),
				},
				&a, &b,
			)

			assert.NoError(t, app.Start(t.Context()))

			assert.NotNil(t, a)
			assert.NotNil(t, b)

			assert.NoError(t, app.Stop(t.Context()))
		})
	})
}
