package main

import (
	"context"
	_ "embed"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/exp/slog"

	"github.com/Dyleme/Notifier/internal/authorization/authmiddleware"
	authHandlerrs "github.com/Dyleme/Notifier/internal/authorization/handler/handlers"
	"github.com/Dyleme/Notifier/internal/authorization/jwt"
	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
	authenticationService "github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/config"
	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/log/slogpretty"
	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
	cmd_notifier "github.com/Dyleme/Notifier/internal/notification-service/cmdnotifier"
	"github.com/Dyleme/Notifier/internal/notification-service/notifier"
	"github.com/Dyleme/Notifier/internal/server"
	"github.com/Dyleme/Notifier/internal/server/custmidlleware"
	timetableHandler "github.com/Dyleme/Notifier/internal/timetable-service/handler/handlers"
	timetableRepository "github.com/Dyleme/Notifier/internal/timetable-service/repository"
	timetableService "github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func main() {
	cfg, err := config.Load()
	logger := setupLogger(cfg.Env)
	if err != nil {
		logger.Error("configuration loading error", log.Err(err))
		return
	}
	ctx := log.InCtx(context.Background(), logger)
	db, err := sqldatabase.NewPGX(ctx, cfg.Database.ConnectionString())
	if err != nil {
		logger.Error("db init error", log.Err(err))
		return
	}

	notif := notifier.New(cmd_notifier.New(logger), cfg.Notifier)
	timetableRepo := timetableRepository.New(db)
	timetableServ := timetableService.New(timetableRepo, notif, cfg.Timetable)
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

	serv := server.New(router, cfg.Server)
	if err := serv.Run(ctx); err != nil {
		logger.Error("server error", log.Err(err))
	}
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
