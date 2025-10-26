package txmanager

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

type loggingDBTX struct {
	DBTX
	ExtractFunc   func(ctx context.Context) *slog.Logger
	IgnoredErrors []error
	ErrorLevel    slog.Level
	LogLevel      slog.Level
}

func WithLogging(extractFunc func(ctx context.Context) *slog.Logger, logLevel, errLevel slog.Level, ignoredErrors []error) Option {
	return func(dbtx DBTX) DBTX {
		return &loggingDBTX{
			DBTX:          dbtx,
			ExtractFunc:   extractFunc,
			IgnoredErrors: ignoredErrors,
			ErrorLevel:    errLevel,
			LogLevel:      logLevel,
		}
	}
}

func (l *loggingDBTX) errorLog(ctx context.Context, err error, query string, args ...any) {
	logger := slog.Default()
	if l.ExtractFunc != nil {
		logger = l.ExtractFunc(ctx)
	}

	if err == nil {
		logger.Log(ctx, l.LogLevel, "database query", "query", query, "args", args)

		return
	}

	for _, ignoredErr := range l.IgnoredErrors {
		if errors.Is(err, ignoredErr) {
			logger.Log(ctx, l.LogLevel, "database query", "query", query, "args", args)

			return
		}
	}

	logger.Log(ctx, l.ErrorLevel, "database error", "error", err, "query", query, "args", args)
}

func (l *loggingDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := l.DBTX.ExecContext(ctx, query, args...)
	if err != nil {
		l.errorLog(ctx, err, query, args...)
	}

	return res, err
}

func (l *loggingDBTX) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := l.DBTX.QueryContext(ctx, query, args...)
	if err != nil {
		l.errorLog(ctx, err, query, args...)
	}

	return rows, err
}

func (l *loggingDBTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row := l.DBTX.QueryRowContext(ctx, query, args...)
	if err := row.Err(); err != nil {
		l.errorLog(ctx, err, query, args...)
	}

	return row
}

func (l *loggingDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	stmt, err := l.DBTX.PrepareContext(ctx, query)
	if err != nil {
		l.errorLog(ctx, err, query)
	}

	return stmt, err
}
