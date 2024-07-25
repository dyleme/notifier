package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	timetableService "github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
)

type Config struct {
	Token string
}

type TelegramHandler struct {
	bot                 *bot.Bot
	serv                *timetableService.Service
	userRepo            UserRepo
	kvRepo              KVRepo
	waitingActionsStore WaitingActionsStore
}

func New(service *timetableService.Service, userRepo UserRepo, cfg Config, actionsStore WaitingActionsStore, kvStore KVRepo) (*TelegramHandler, error) {
	op := "New: %w"
	tgHandler := TelegramHandler{
		kvRepo:              kvStore,
		serv:                service,
		userRepo:            userRepo,
		waitingActionsStore: actionsStore,
		bot:                 nil, // set this field later by calling SetBot method
	}
	opts := []bot.Option{
		bot.WithMiddlewares(recoverPanicMiddleware, loggingMiddleware, tgHandler.UserMiddleware),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, tgHandler.StartListener),
		bot.WithMessageTextHandler("/info", bot.MatchTypeExact, tgHandler.InfoListener),
		bot.WithMessageTextHandler("/cancel", bot.MatchTypeExact, tgHandler.CancelListener),
		bot.WithDefaultHandler(tgHandler.Handle),
	}

	b, err := bot.New(cfg.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tgHandler.bot = b

	return &tgHandler, nil
}

func (th *TelegramHandler) Run(ctx context.Context) {
	log.Ctx(ctx).Info("start telegram bot")
	th.bot.Start(ctx)
}

type TextMessageHandler struct {
	handle    func(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error
	messageID int
}

type WaitingActionsStore interface {
	StoreDefDur(key int64, val TextMessageHandler)
	Get(key int64) (TextMessageHandler, error)
	Delete(key int64)
}

type UserRepo interface {
	GetUserInfo(ctx context.Context, tgID int) (userinfo.User, error)
	UpdateUserTime(ctx context.Context, tgID int, timezoneOffset int, isDST bool) error
}

type KVRepo interface {
	PutValue(ctx context.Context, key string, value any) error
	GetValue(ctx context.Context, key string, value any) error
	DeleteValue(ctx context.Context, key string) error
}

func (th *TelegramHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	op := "TelegramHandler.Handle: %w"
	chatID, err := th.chatID(update)
	if err != nil {
		handleError(ctx, b, 0, err)

		return
	}
	if update.Message != nil {
		textHandler, err := th.waitingActionsStore.Get(chatID)
		if err != nil {
			if creationError := th.mainMenuCreateWindow(ctx, b, chatID); creationError != nil {
				handleError(ctx, b, chatID, fmt.Errorf(op, creationError))
			}

			return
		}

		err = textHandler.handle(ctx, b, update.Message, textHandler.messageID)
		if err != nil {
			handleError(ctx, b, chatID, fmt.Errorf(op, err))
		}

		return
	}

	if creationError := th.mainMenuCreateWindow(ctx, b, chatID); creationError != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, creationError))
	}
}

func (th *TelegramHandler) InfoListener(ctx context.Context, b *bot.Bot, update *models.Update) {
	op := "TelegramHandler.InfoListener: %w"

	if update.Message != nil {
		if err := th.InfoInline(ctx, b, update.Message, nil); err != nil {
			handleError(ctx, b, update.Message.Chat.ID, fmt.Errorf(op, err))
		}
	}
}

func (th *TelegramHandler) StartListener(ctx context.Context, b *bot.Bot, update *models.Update) {
	op := "TelegramHandler.StartListener: %w"
	chatID, err := th.chatID(update)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}

	err = th.mainMenuCreateWindow(ctx, b, chatID)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}
}

func (th *TelegramHandler) CancelListener(ctx context.Context, b *bot.Bot, update *models.Update) {
	op := "TelegramHandler.CancelListener: %w"
	chatID, err := th.chatID(update)
	if err != nil {
		handleError(ctx, b, 0, fmt.Errorf(op, err))

		return
	}

	th.waitingActionsStore.Delete(chatID)

	_, err = th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Return basic state",
	})
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}

	err = th.mainMenuCreateWindow(ctx, b, chatID)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}
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

	return 0, errors.New("no chat id")
}
