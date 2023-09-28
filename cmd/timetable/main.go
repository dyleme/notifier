package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/Dyleme/timecache"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	authHandlerrs "github.com/Dyleme/Notifier/internal/authorization/handler/handlers"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authenticationService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/log/slogpretty"
	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
	"github.com/Dyleme/Notifier/internal/notification-service/cmdnotifier"
	"github.com/Dyleme/Notifier/internal/notification-service/notifier"
	"github.com/Dyleme/Notifier/internal/server"
	"github.com/Dyleme/Notifier/internal/server/custmidlleware"
	"github.com/Dyleme/Notifier/internal/telegram/handler"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	timetableHandler "github.com/Dyleme/Notifier/internal/timetable-service/handler/handlers"
	timetableRepository "github.com/Dyleme/Notifier/internal/timetable-service/repository"
	timetableService "github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func main() { //nolint:funlen // main can be long
	cfg, err := config.Load()
	logger := setupLogger(cfg.Env)
	if err != nil {
		logger.Error("configuration loading error", log.Err(err))

		return
	}
	ctx := log.InCtx(context.Background(), logger)
	ctx = cancelOnInterruption(ctx)

	db, err := sqldatabase.NewPGX(ctx, cfg.Database.ConnectionString())
	if err != nil {
		logger.Error("db init error", log.Err(err))

		return
	}

	notif := notifier.New(ctx, cmdnotifier.New(logger), cfg.Notifier)
	cache := timetableRepository.NewUniversalCache()
	timetableRepo := timetableRepository.New(db, cache)
	timetableServ := timetableService.New(ctx, timetableRepo, notif, cfg.Event)
	timeTableHandler := timetableHandler.New(timetableServ)

	apiTokenMiddleware := authmiddleware.NewAPIToken(cfg.APIKey)
	jwtGen := jwt.NewJwtGen(cfg.JWT)
	jwtMiddleware := authmiddleware.NewJWT(jwtGen)
	authRepo := authRepository.New(db)
	authService := authenticationService.NewAuth(authRepo, &authenticationService.HashGen{}, jwtGen)
	authHandler := authHandlerrs.New(authService)

	router := server.Route(
		timeTableHandler,
		authHandler,
		jwtMiddleware.Handle,
		apiTokenMiddleware.Handle,
		[]func(next http.Handler) http.Handler{
			cors.AllowAll().Handler,
			custmidlleware.WithLogger(logger),
			custmidlleware.RequestID,
			custmidlleware.LoggerMiddleware,
			middleware.Recoverer,
		},
	)

	tg, err := handler.New(timetableServ, userinfo.NewUserRepoCache(authService), cfg.Telegram, timecache.New[int64, handler.TextMessageHandler]())
	if err != nil {
		logger.Error("tg init error", log.Err(err))

		return
	}

	notif.SetNotifier(tg)

	go timetableServ.RunJob(ctx)

	serv := server.New(router, cfg.Server)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := serv.Run(ctx); err != nil { //nolint:govet //new error
			return fmt.Errorf("server: %w", err)
		}

		return nil
	})
	wg.Go(func() error {
		tg.Run(ctx)

		return nil
	})
	err = wg.Wait()
	if err != nil {
		logger.Error("serve error", log.Err(err))
	}
}

func cancelOnInterruption(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		cancel()
	}()

	return ctx
}

const (
	localEnv = "local"
	devEnv   = "dev"
	prodEnv  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case localEnv:
		prettyHandler := slogpretty.NewHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct //no need to set this params
		logger = slog.New(prettyHandler)
	case devEnv:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})) //nolint:exhaustruct //no need to set this params
	case prodEnv:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
	}

	return logger
}
