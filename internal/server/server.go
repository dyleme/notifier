package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

// Server is a struct which handles the requests.
type Server struct {
	server               *http.Server
	gracefulShutdownTime time.Duration
}

type Config struct {
	Port                    int
	MaxHeaderBytes          int
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	TimeForGracefulShutdown time.Duration
}

func New(handler http.Handler, cfg *Config) Server {
	return Server{
		server: &http.Server{ //nolint:exhaustruct //TODO:implement everything
			Addr:           ":" + strconv.Itoa(cfg.Port),
			Handler:        handler,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
		},
		gracefulShutdownTime: cfg.TimeForGracefulShutdown,
	}
}

// After Run method Server starts to listen port and response to  the reqeusts.
// Run function provide the abitility of the gracefule shutdown.
func (s *Server) Run(ctx context.Context) error {
	servError := make(chan error, 1)

	go func() {
		log.Ctx(ctx).Info("start server")
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			servError <- fmt.Errorf("listen: %w", err)
		}
	}()

	select {
	case err := <-servError:
		log.Ctx(ctx).Error("server error", log.Err(err))

		return err
	case <-ctx.Done():
		gsStart := time.Now()
		log.Ctx(ctx).Info("server start graceful shutdown")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTime)
		defer cancel()

		if err := s.server.Shutdown(ctxShutDown); err != nil { //nolint:contextcheck //create new context for graceful shutdown
			return fmt.Errorf("shutdown: %w", err)
		}

		log.Ctx(ctx).Info("server end graceful shutdown", slog.Duration("shutdown_dur", time.Since(gsStart)))
	}

	return nil
}
