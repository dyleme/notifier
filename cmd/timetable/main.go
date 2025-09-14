package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/Dyleme/timecache"
	"github.com/benbjohnson/clock"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"

	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/httpserver"
	custMiddleware "github.com/Dyleme/Notifier/internal/httpserver/middleware"
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/service/handler"
	"github.com/Dyleme/Notifier/internal/service/repository"
	"github.com/Dyleme/Notifier/internal/service/service"
	tgHandler "github.com/Dyleme/Notifier/internal/telegram/handler"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/database/sqldatabase"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/jobontime"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/slogpretty"
)

func main() { //nolint:funlen // main can be long
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := setupLogger(cfg.Env)
	ctx := log.InCtx(context.Background(), logger)
	ctx = cancelOnInterruption(ctx)

	db, err := sqldatabase.NewSQLite(ctx, cfg.DatabaseFile)
	if err != nil {
		logger.Error("db init error", log.Err(err))

		return
	}

	cache := repository.NewUniversalCache()
	txManager := txmanager.New(db)
	txGetter := txmanager.NewGetter(db)
	nower := clock.New()
	eventsNotifier := eventnotifier.New(
		repository.NewEventsRepository(txGetter),
		txManager,
	)
	eventsNotifierJob := jobontime.New(
		nower,
		eventsNotifier,
		cfg.NotifierJob.CheckTasksPeriod,
	)
	svc := service.New(
		repository.NewPeriodicTaskRepository(txGetter),
		repository.NewBasicTaskRepository(txGetter),
		repository.NewTGImagesRepository(txGetter, cache),
		repository.NewEventsRepository(txGetter),
		repository.NewDefaultNotificationParamsRepository(txGetter),
		repository.NewTagsRepository(txGetter),
		trManager,
		eventsNotifierJob,
	)
	timeTableHndlr := handler.New(svc)

	authRepo := authRepository.New(db, txGetter)
	authSvc := authService.NewAuth(
		authRepo,
		&authService.HashGen{},
		jwtGen,
		trManager,
		codeGen,
	)

	router := httpserver.Route(
		timeTableHndlr,
		authHndlr,
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

	tg, err := tgHandler.New(
		svc,
		userinfo.NewUserRepoCache(authSvc),
		cfg.Telegram,
		timecache.New[int64, tgHandler.TextMessageHandler](),
		repository.NewKeyValueRepository(db, txGetter),
	)
	if err != nil {
		logger.Error("tg init error", log.Err(err))

		return
	}

	eventsNotifier.SetNotifier(tg)
	authSvc.SetCodeSender(tg)

	go eventsNotifierJob.Run(ctx)

	serv := httpserver.New(router, cfg.Server)

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
