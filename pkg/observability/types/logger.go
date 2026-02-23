package obs_types

import "context"

//mockery:generate: true
type Logger interface {
	Info(ctx context.Context, msg string, attributes ...any)
	Error(ctx context.Context, msg string, err error, attributes ...any)
	// CtxWithLogAttributes returns a new context with the given attributes injected, it is
	// expected that these attributes will be included in all subsequent logs made with this context
	CtxWithLogAttributes(ctx context.Context, attributes ...any) context.Context
}
