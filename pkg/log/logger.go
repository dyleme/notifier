package log

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type key string

const ctxKey key = "contextKey"

func Ctx(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(ctxKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}

func InCtx(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}

func WithCtx(ctx context.Context, args ...any) context.Context {
	logger := Ctx(ctx)
	logger = logger.With(args...)

	return InCtx(ctx, logger)
}

func RequestID() slog.Attr {
	return slog.Attr{
		Key:   "request_id",
		Value: slog.StringValue(uuid.NewString()),
	}
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Default() *slog.Logger {
	return slog.Default()
}
