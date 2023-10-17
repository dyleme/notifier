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
	authorizatoinHandler "github.com/Dyleme/Notifier/internal/authorization/handler"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authorizationRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authorizationService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/notifier"
	"github.com/Dyleme/Notifier/internal/server"
	custMiddleware "github.com/Dyleme/Notifier/internal/server/middleware"
	timetableHandler "github.com/Dyleme/Notifier/internal/service/handler"
	timetableRepository "github.com/Dyleme/Notifier/internal/service/repository"
	timetableService "github.com/Dyleme/Notifier/internal/service/service"
	tgHandler "github.com/Dyleme/Notifier/internal/telegram/handler"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/slogpretty"
	"github.com/Dyleme/Notifier/pkg/sqldatabase"
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

	notif := notifier.New(ctx, nil, cfg.Notifier)
	cache := timetableRepository.NewUniversalCache()
	repo := timetableRepository.New(db, cache)
	service := timetableService.New(ctx, repo, notif, cfg.Service)
	timeTableHandler := timetableHandler.New(service)

	apiTokenMiddleware := authmiddleware.NewAPIToken(cfg.APIKey)
	jwtGen := jwt.NewJwtGen(cfg.JWT)
	jwtMiddleware := authmiddleware.NewJWT(jwtGen)
	authRepo := authorizationRepository.New(db)
	authService := authorizationService.NewAuth(authRepo, &authorizationService.HashGen{}, jwtGen)
	authHandler := authorizatoinHandler.New(authService)

	router := server.Route(
		timeTableHandler,
		authHandler,
		jwtMiddleware.Handle,
		apiTokenMiddleware.Handle,
		[]func(next http.Handler) http.Handler{
			cors.AllowAll().Handler,
			custMiddleware.WithLogger(logger),
			custMiddleware.LoggerMiddleware,
			custMiddleware.RequestID,
			middleware.Recoverer,
		},
	)

	tg, err := tgHandler.New(service, userinfo.NewUserRepoCache(authService), cfg.Telegram, timecache.New[int64, tgHandler.TextMessageHandler]())
	if err != nil {
		logger.Error("tg init error", log.Err(err))

		return
	}

	notif.SetNotifier(tg)

	go service.RunNotificationJob(ctx)

	serv := server.New(router, cfg.Server)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := serv.Run(ctx); err != nil {
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
