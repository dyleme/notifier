package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/sync/errgroup"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	authHandler "github.com/Dyleme/Notifier/internal/authorization/handler"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/notifierjob"
	"github.com/Dyleme/Notifier/internal/server"
	custMiddleware "github.com/Dyleme/Notifier/internal/server/middleware"
	"github.com/Dyleme/Notifier/internal/service/handler"
	"github.com/Dyleme/Notifier/internal/service/repository"
	"github.com/Dyleme/Notifier/internal/service/service"
	tgHandler "github.com/Dyleme/Notifier/internal/telegram/handler"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/slogpretty"
	"github.com/Dyleme/Notifier/pkg/sqldatabase"
	"github.com/Dyleme/timecache"
)

func main() { //nolint:funlen // main can be long
	cfg := config.Load()
	logger := setupLogger(cfg.Env)
	ctx := log.InCtx(context.Background(), logger)
	ctx = cancelOnInterruption(ctx)

	db, err := sqldatabase.NewPGX(ctx, cfg.Database.ConnectionString())
	if err != nil {
		logger.Error("db init error", log.Err(err))

		return
	}

	cache := repository.NewUniversalCache()
	trManager := manager.Must(trmpgx.NewDefaultFactory(db))
	trCtxGetter := trmpgx.DefaultCtxGetter
	notifierJob := notifierjob.New(
		repository.NewEventsRepository(db, trCtxGetter),
		cfg.NotifierJob,
		trManager,
	)
	svc := service.New(
		repository.NewPeriodicTaskRepository(db, trCtxGetter),
		repository.NewBasicTaskRepository(db, trCtxGetter),
		repository.NewTGImagesRepository(db, trCtxGetter, cache),
		repository.NewEventsRepository(db, trCtxGetter),
		repository.NewDefaultNotificationParamsRepository(db, trCtxGetter),
		trManager,
		notifierJob,
	)
	timeTableHndlr := handler.New(svc)

	apiTokenMiddleware := authmiddleware.NewAPIToken(cfg.APIKey)
	jwtGen := jwt.NewJwtGen(cfg.JWT)
	codeGen := authService.NewRandomIntSeq()
	jwtMiddleware := authmiddleware.NewJWT(jwtGen)
	authRepo := authRepository.New(db, trCtxGetter)
	authSvc := authService.NewAuth(
		authRepo,
		&authService.HashGen{},
		jwtGen,
		trManager,
		codeGen,
	)
	authHndlr := authHandler.New(authSvc)

	router := server.Route(
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
		repository.NewKeyValueRepository(db, trCtxGetter),
	)
	if err != nil {
		logger.Error("tg init error", log.Err(err))

		return
	}

	notifierJob.SetNotifier(notifierjob.CmdNotifier{})
	authSvc.SetCodeSender(authService.CmdCodeSender{})

	go notifierJob.Run(ctx)

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
