package util

import (
	"context"

	"go.uber.org/fx"
)

// Lifeycle - On start and on stop hooks for a module. These can be useful if the module has to manage some
// sort of state, e.g. a db connection pool.
type FxLifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// New - Helper for creating a module for the fx dependency framework. T is the interface
// being added to the dependency graph. The constructor must return a concrete implementation of T.
// The constructor may require other components of the dependency graph, which will be resolved by fx.
//
// opts - additional fx options to add to this module
func NewFxModule[T any](
	name string,
	constructor any,
	opts ...fx.Option) fx.Option {
	return fx.Module(
		name,
		append(
			[]fx.Option{
				fx.Provide(
					fx.Annotate(constructor, fx.As(new(T))),
				),
			},
			opts...,
		)...,
	)
}

// NewWithLifeCycle - similar to New, but for modules with lifecycle hooks,
// it is expected the module implements FxLifecycle
func NewFxModuleWithLifecycle[T FxLifecycle](
	name string,
	constructor any,
	opts ...fx.Option,
) fx.Option {
	return NewFxModule[T](
		name,
		constructor,
		append(
			opts,
			fx.Invoke(registerFxLifecycle[T]),
		)...,
	)
}

// CreateFxAppAndExtract creates an Fx application and extracts specified dependencies
// to be used out of the Fx application context. This should be done sparingly, at the edge
// of the system. E.g. Pulling out an HTTP server to run on the main thread, vs
// letting Fx run it in its own goroutine. Fx internal logging is disabled by default.
func CreateFxAppAndExtract(modules []fx.Option, extract ...any) *fx.App {
	return fx.New(
		append(
			modules,
			// disable fx logging
			fx.NopLogger,
			fx.Populate(extract...),
		)...,
	)
}

// registerLifecycle - wires hooks for component start and stop
func registerFxLifecycle[T FxLifecycle](lc fx.Lifecycle, this T) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return this.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return this.Stop(ctx)
		},
	})
}
