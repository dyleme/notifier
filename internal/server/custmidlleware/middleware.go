package custmidlleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

func WithLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := log.InCtx(r.Context(), logger)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := log.Ctx(ctx)
		logger = logger.With(slog.String("req_id", uuid.NewString()))
		r = r.WithContext(log.InCtx(ctx, logger))
		next.ServeHTTP(w, r)
	})
}

func LoggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())
		entry := logger.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			entry.Info("request completed",
				slog.Int("status", ww.Status()),
				slog.String("duration", time.Since(t1).String()),
			)
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
