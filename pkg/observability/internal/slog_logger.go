package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

type attributesContextKey struct{}

type slogger struct {
	logger *slog.Logger
}

func NewSlogLogger(cfg obs_types.SlogLoggerConfig) (obs_types.Logger, error) {
	var level slog.Level
	switch cfg.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("unknown log level %v", cfg.Level)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	return &slogger{
		logger: logger,
	}, nil
}

func (l *slogger) Info(ctx context.Context, msg string, attributes ...any) {
	l.logger.Info(msg, append(attributesFromContext(ctx), attributes...)...)
}

func (l *slogger) Error(ctx context.Context, msg string, err error, attributes ...any) {
	finalAttrs := append(attributesFromContext(ctx), attributes...)
	finalAttrs = append(finalAttrs, "error", err)
	l.logger.Error(msg, finalAttrs...)
}

func (l *slogger) CtxWithLogAttributes(ctx context.Context, attributes ...any) context.Context {
	existingAttrs, _ := ctx.Value(attributesContextKey{}).([]any)
	newAttrs := make([]any, 0, len(existingAttrs)+len(attributes))
	newAttrs = append(newAttrs, existingAttrs...)
	newAttrs = append(newAttrs, attributes...)
	return context.WithValue(ctx, attributesContextKey{}, newAttrs)
}

// attributesFromContext extracts attributes from the context, such as trace ids, user ids, etc, that
// may be injected into the context by middleware
func attributesFromContext(ctx context.Context) []any {
	attrs, _ := ctx.Value(attributesContextKey{}).([]any)
	return attrs
}
