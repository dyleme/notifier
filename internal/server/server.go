package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"golang.org/x/exp/slog"

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
		server: &http.Server{
			Addr:           ":" + strconv.Itoa(cfg.Port),
			Handler:        handler,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
		},
		gracefulShutdownTime: cfg.TimeForGracefulShutdown,
	}
}

func catchOSInterrupt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		cancel()
	}()
}

// After Run method Server starts to listen port and response to  the reqeusts.
// Run function provide the abitility of the gracefule shutdown.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	catchOSInterrupt(cancel)

	servError := make(chan error, 1)

	go func() {
		log.Ctx(ctx).Info("start server")
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			servError <- fmt.Errorf("listen: %w", err)
		}
	}()

	select {
	case err := <-servError:
		return err
	case <-ctx.Done():
		gsStart := time.Now()
		log.Ctx(ctx).Info("starting graceful shutdown")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTime)
		defer cancel()

		if err := s.server.Shutdown(ctxShutDown); err != nil { // nolint: contextcheck // create new context for graceful shutdown
			return fmt.Errorf("shutdown: %w", err)
		}

		log.Ctx(ctx).Info("end graceful shutdown", slog.Duration("shutdown_dur", time.Since(gsStart)))
	}

	return nil
}
