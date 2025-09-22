package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Dyleme/timecache"
	"github.com/benbjohnson/clock"

	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/Dyleme/Notifier/internal/repository"
	"github.com/Dyleme/Notifier/internal/service"
	"github.com/Dyleme/Notifier/internal/telegram"
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
	eventsNotifier := eventnotifier.New(txManager)
	eventsNotifierJob := jobontime.New(
		nower,
		eventsNotifier,
		cfg.NotifierJob.CheckTasksPeriod,
	)
	userRepo := repository.NewUserRepository(txGetter)

	svc := service.New(
		userRepo,
		repository.NewTasksRepository(txGetter),
		repository.NewTGImagesRepository(txGetter, cache),
		repository.NewEventsRepository(txGetter),
		txManager,
		eventsNotifierJob,
	)

	tg, err := telegram.New(
		svc,
		cfg.Telegram,
		timecache.New[int64, telegram.TextMessageHandler](),
		repository.NewKeyValueRepository(txGetter),
	)
	if err != nil {
		logger.Error("tg init error", log.Err(err))

		return
	}

	eventsNotifier.SetNotifier(tg)
	eventsNotifier.SetService(svc)

	go eventsNotifierJob.Run(ctx)

	tg.Run(ctx)
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
