package mocklogger

import (
	"context"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
)

type MockHandler struct {
	slog.Handler
	l   *stdLog.Logger
	err error
}

func NewHandler(
	out io.Writer,
	opts *slog.HandlerOptions,
) *MockHandler {
	h := &MockHandler{
		Handler: slog.NewJSONHandler(out, opts),
		l:       stdLog.New(out, "", 0),
		err:     nil,
	}

	return h
}

func (h *MockHandler) Error() error {
	return h.err
}

func (h *MockHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level == slog.LevelError {
		if h.err == nil {
			fields := make(map[string]interface{}, r.NumAttrs())

			r.Attrs(func(a slog.Attr) bool {
				fields[a.Key] = a.Value.Any()

				return true
			})

			h.err = fmt.Errorf("logged error: %s %v", r.Message, fields)
		}
	}

	return nil
}

func (h *MockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockHandler) WithGroup(_ string) slog.Handler {
	return h
}
