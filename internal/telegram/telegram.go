package main

//
// import (
// 	"context"
// 	"log/slog"
// 	"os"
// 	"os/signal"
//
// 	"github.com/go-telegram/bot"
//
// 	authRepository "github.com/Dyleme/Notifier/internal/authorization/repository"
// 	"github.com/Dyleme/Notifier/internal/config"
// 	"github.com/Dyleme/Notifier/internal/lib/log"
// 	"github.com/Dyleme/Notifier/internal/lib/log/slogpretty"
// 	"github.com/Dyleme/Notifier/internal/lib/sqldatabase"
// 	"github.com/Dyleme/Notifier/internal/notification-service/cmdnotifier"
// 	"github.com/Dyleme/Notifier/internal/notification-service/notifier"
// 	"github.com/Dyleme/Notifier/internal/telegram/handler"
// 	timetableRepository "github.com/Dyleme/Notifier/internal/timetable-service/repository"
// 	timetableService "github.com/Dyleme/Notifier/internal/timetable-service/service"
// )
//
// type Server struct {
// 	userIDRepo UserIDRepository
// 	service    timetableService.Service
// 	tgHandler  handler.TelegramHandler
// }
//
// func (s *Server) Run(ctx context.Context) {
// 	s.tgHandler.SetBot(b)
//
// 	b.Start(ctx)
// }
//
// func New(userIDRepo UserIDRepository, serv timetableService.Service) *Server {
//
// 	b, err := bot.New("6571169482:AAGOjtRiHqW7rR8xYwcWplGrkxneEfWpMT8", opts...)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return &Server{
// 		userIDRepo: userIDRepo,
// 		service:    serv,
// 	}
// }
//
// type UserIDRepository interface {
// 	GetID(ctx context.Context, tgID, tgChatID int) (userID int, err error)
// }
//
// // Send any text message to the bot after the bot has been started
// func main() {
// 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer cancel()
//
// 	cfg, err := config.Load()
// 	logger := setupLogger(cfg.Env)
// 	if err != nil {
// 		logger.Error("configuration loading error", log.Err(err))
// 		return
// 	}
//
// 	ctx = log.InCtx(ctx, logger)
//
// 	db, err := sqldatabase.NewPGX(ctx, cfg.Database.ConnectionString())
// 	if err != nil {
// 		logger.Error("db init error", log.Err(err))
// 		return
// 	}
//
// 	notif := notifier.New(ctx, cmdnotifier.New(logger), cfg.Notifier)
// 	timetableRepo := timetableRepository.New(db)
// 	timetableServ := timetableService.New(ctx, timetableRepo, notif, cfg.Timetable)
// 	go timetableServ.RunJob(ctx)
//
// 	authRepo := authRepository.New(db)
// 	userIDCacheRepo := handler.NewUserRepoCache(authRepo)
// 	tgHandler := handler.New(timetableServ, userIDCacheRepo)
//
// 	notif.SetNotifier(tgHandler)
//
// }
//
// const (
// 	localEnv = "local"
// 	devEnv   = "dev"
// 	prodEnv  = "prod"
// )
//
// func setupLogger(env string) *slog.Logger {
// 	var logger *slog.Logger
//
// 	switch env {
// 	case localEnv:
// 		prettyHandler := slogpretty.NewHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct //no need to set this params
// 		logger = slog.New(prettyHandler)
// 	case devEnv:
// 		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})) //nolint:exhaustruct //no need to set this params
// 	case prodEnv:
// 		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
// 	default:
// 		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})) //nolint:exhaustruct //no need to set this params
// 	}
//
// 	return logger
// }
