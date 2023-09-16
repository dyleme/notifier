package slogpretty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"

	"github.com/fatih/color"
)

type PrettyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func NewHandler(
	out io.Writer,
	opts *slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts),
		l:       stdLog.New(out, "", 0),
		attrs:   nil,
	}

	return h
}

const (
	errField      = "error"
	callPathField = "callPath"
)

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	op := "PrettyHandler.Handle: %w"
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs()+len(h.attrs))

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var errMsg string

	if val, ok := fields[errField]; ok {
		errMsg, ok = val.(string)
		if ok {
			delete(fields, errField)
		}
	}

	var fieldsMsg string
	if len(fields) > 0 {
		fieldsBytes, err := json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return fmt.Errorf(op, err)
		}
		fieldsMsg = "\n" + string(fieldsBytes)
	}

	timeStr := r.Time.Format("[15:05:05.000]")

	h.l.Println(
		timeStr,
		level,
		color.CyanString(r.Message),
		color.RedString(errMsg),
		color.WhiteString(fieldsMsg),
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
		attrs:   nil,
	}
}
