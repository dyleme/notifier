package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/tgwf"
	timetableService "github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type Config struct {
	Token string
}

const defaultListLimit = 100

func New(service *timetableService.Service, userRepo UserRepo, cfg Config) (*TelegramHandler, error) {
	op := "New: %w"
	tgHandler := TelegramHandler{
		serv:     service,
		userRepo: userRepo,
		bot:      nil, // set this field later by calling SetBot method
	}
	opts := []bot.Option{
		bot.WithMiddlewares(loggingMiddleware, tgHandler.UserIDMiddleware),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, tgHandler.MainMenu),
		bot.WithMessageTextHandler("/info", bot.MatchTypeExact, tgHandler.Info),
		bot.WithMessageTextHandler("/cancel", bot.MatchTypeExact, tgHandler.Cancel),
		bot.WithDefaultHandler(tgHandler.Handle),
	}

	b, err := bot.New(cfg.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tgHandler.bot = tgwf.New(b, tgHandler.MainMenuAction())

	return &tgHandler, nil
}

type TelegramHandler struct {
	bot      *tgwf.WorkflowHandler
	serv     *timetableService.Service
	userRepo UserRepo
}

func (th *TelegramHandler) Run(ctx context.Context) {
	log.Ctx(ctx).Info("start telegram bot")
	th.bot.Bot.Start(ctx)
}

type UserRepo interface {
	GetID(ctx context.Context, tgID int) (userID int, err error)
}

func (th *TelegramHandler) SetBot(b *bot.Bot) {
	th.bot = tgwf.New(b, th.MainMenuAction())
}

func (th *TelegramHandler) tgUserID(update *models.Update) (int64, error) {
	switch {
	case update.Message != nil:
		return update.Message.From.ID, nil
	case update.CallbackQuery != nil:
		return update.CallbackQuery.Sender.ID, nil
	}

	return 0, fmt.Errorf("unknown id")
}

func (th *TelegramHandler) chatID(update *models.Update) (int64, error) {
	switch {
	case update.Message != nil:
		return update.Message.Chat.ID, nil
	case update.CallbackQuery != nil:
		if update.CallbackQuery.Message != nil {
			return update.CallbackQuery.Message.Chat.ID, nil
		}
	}

	return 0, fmt.Errorf("no chat id")
}

func (th *TelegramHandler) Handle(ctx context.Context, _ *bot.Bot, update *models.Update) {
	chatID, err := th.chatID(update)
	if err != nil {
		th.handleError(ctx, 0, err)

		return
	}
	err = th.bot.HandleAction(ctx, update)
	if err != nil {
		if errors.Is(err, tgwf.ErrNoAssociatedAction) {
			th.Info(ctx, nil, update)

			return
		}
		th.handleError(ctx, chatID, err)

		return
	}
}

func (th *TelegramHandler) Info(ctx context.Context, _ *bot.Bot, update *models.Update) {
	chatID, err := th.chatID(update)
	if err != nil {
		th.handleError(ctx, 0, err)

		return
	}

	err = th.info(ctx, chatID)
	if err != nil {
		th.handleError(ctx, chatID, err)

		return
	}
}

func (th *TelegramHandler) MainMenu(ctx context.Context, _ *bot.Bot, update *models.Update) {
	op := "TelegramHandler.MainMenu: %w"
	chatID, err := th.chatID(update)
	if err != nil {
		handleError(ctx, th.bot, chatID, fmt.Errorf(op, err))

		return
	}

	err = th.bot.Start(ctx, chatID, th.MainMenuAction())
	if err != nil {
		handleError(ctx, th.bot, chatID, fmt.Errorf(op, err))

		return
	}
}

func (th *TelegramHandler) MainMenuAction() tgwf.Action {
	menu := tgwf.NewMenuAction("What do you want to do?").
		Row().Btn("Info", th.ShowInfo).
		Row().Btn("Tasks", th.TaskMenu()).
		Row().Btn("Events", th.EventsMenu()).
		Row().Btn("Notifications", th.NotificationMenu())

	return menu.Show
}

func (th *TelegramHandler) mainMenu(ctx context.Context, chatID int64) error {
	op := "TelegramHandler.menu: %w"

	menu := th.MainMenuAction()
	err := th.bot.Start(ctx, chatID, menu)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) Cancel(ctx context.Context, _ *bot.Bot, update *models.Update) {
	op := "TelegramHandler.Cancel: %w"
	chatID, err := th.chatID(update)
	if err != nil {
		th.handleError(ctx, 0, fmt.Errorf(op, err))

		return
	}

	th.bot.ForgotForChat(ctx, chatID)

	_, err = th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Return basic state",
	})
	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}

	err = th.mainMenu(ctx, chatID)
	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}
}

func (th *TelegramHandler) handleError(ctx context.Context, chatID int64, err error) {
	handleError(ctx, th.bot, chatID, err)
}

func handleError(ctx context.Context, b *tgwf.WorkflowHandler, chatID int64, err error) {
	log.Ctx(ctx).Error("error occurred", log.Err(err))
	if chatID == 0 {
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Server error occurred",
	})
	if err != nil {
		log.Ctx(ctx).Error("cannot send error message", log.Err(err))
	}
}
