package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/Dyleme/timecache"
	"github.com/benbjohnson/clock"

	"github.com/dyleme/Notifier/internal/config"
	"github.com/dyleme/Notifier/internal/notifier/eventnotifier"
	"github.com/dyleme/Notifier/internal/repository"
	"github.com/dyleme/Notifier/internal/service"
	"github.com/dyleme/Notifier/internal/telegram"
	"github.com/dyleme/Notifier/pkg/database/sqldatabase"
	"github.com/dyleme/Notifier/pkg/database/txmanager"
	"github.com/dyleme/Notifier/pkg/jobontime"
	"github.com/dyleme/Notifier/pkg/log"
	"github.com/dyleme/Notifier/pkg/log/slogpretty"
)

func main() { //nolint:funlen // main can be long
	var closeFuncs []func() error
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, closeLogFile, err := setupLogger(cfg.Env, cfg.LogFile)
	if err != nil {
		panic(err)
	}
	closeFuncs = append(closeFuncs, closeLogFile)
	defer func() {
		for i := len(closeFuncs) - 1; i >= 0; i-- {
			if err := closeFuncs[i](); err != nil {
				logger.Error("close func error", log.Err(err))
			}
		}
	}()

	ctx := log.InCtx(context.Background(), logger)
	ctx = cancelOnInterruption(ctx)

	db, closeDB, err := sqldatabase.NewSQLite(ctx, cfg.DatabaseFile)
	if err != nil {
		logger.Error("db init error", log.Err(err))

		return
	}
	closeFuncs = append(closeFuncs, closeDB)

	cache := repository.NewUniversalCache()
	txManager, txGetter := txmanager.New(db, txmanager.WithLogging(log.Ctx, txmanager.LoggingSetting{
		LogLevel:   slog.LevelDebug,
		ErrorLevel: slog.LevelError,
	}))
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

	run(
		ctx,
		[]func(ctx context.Context){
			eventsNotifierJob.Run,
			tg.Run,
		},
	)
}

func run(ctx context.Context, funcs []func(ctx context.Context)) {
	var wg sync.WaitGroup
	for _, f := range funcs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			f(ctx)
		}()
	}

	wg.Wait()
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

	ownerWriteOthersRead = 0o644
)

func setupLogger(env, logFile string) (*slog.Logger, func() error, error) {
	out := os.Stdout

	if logFile != "" {
		var err error
		out, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, ownerWriteOthersRead)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
	}

	var logger *slog.Logger
	switch env {
	case localEnv:
		prettyHandler := slogpretty.NewHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct //no need to set this params
		logger = slog.New(prettyHandler)
	case devEnv:
		logger = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})) //nolint:exhaustruct //no need to set this params
	case prodEnv:
		logger = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
	default:
		logger = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
	}

	return logger, out.Close, nil
}
