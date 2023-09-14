package log

import (
	"context"
	"log/slog"
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

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Default() *slog.Logger {
	return slog.Default()
}
