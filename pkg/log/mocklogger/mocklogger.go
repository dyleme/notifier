package mocklogger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type MockHandler struct {
	slog.Handler
	err   error
	errMx sync.Mutex
}

func NewHandler() *MockHandler {
	h := &MockHandler{
		Handler: slog.NewJSONHandler(nil, nil),
		err:     nil,
		errMx:   sync.Mutex{},
	}

	return h
}

func (h *MockHandler) Error() error {
	return h.err
}

func (h *MockHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level != slog.LevelError {
		return nil
	}

	h.errMx.Lock()
	defer h.errMx.Unlock()
	if h.err != nil {
		return nil
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	h.err = fmt.Errorf("logged error: %s %v", r.Message, fields)

	return nil
}

func (h *MockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockHandler) WithGroup(_ string) slog.Handler {
	return h
}
