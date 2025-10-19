package telegram

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/dyleme/Notifier/internal/domain"
	serverrors "github.com/dyleme/Notifier/internal/domain/apperr"
	"github.com/dyleme/Notifier/pkg/log"
)

func chatID(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.EditedMessage != nil {
		return update.EditedMessage.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}

	return 0
}

func recoverPanicMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				handleError(log.WithCtx(ctx, "stack", string(stack)), bot, chatID(update), fmt.Errorf("%v", r))
			}
		}()

		next(ctx, bot, update)
	}
}

const (
	defaultNotificatinPeriod = 5 * time.Minute
)

func (th *Handler) UserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID, err := th.chatID(update)
		if err != nil {
			handleError(ctx, b, 0, err)

			return
		}

		var tgUserID int64
		switch {
		case update.Message != nil:
			tgUserID = update.Message.From.ID
		case update.CallbackQuery != nil:
			tgUserID = update.CallbackQuery.From.ID
		}

		user, err := th.serv.GetTGUser(ctx, int(tgUserID))
		if err != nil {
			var notFoundErr serverrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				handleError(ctx, b, chatID, err)

				return
			}

			user = domain.User{
				TGID:                      int(tgUserID),
				TimeZoneOffset:            0,
				IsTimeZoneDST:             false,
				DefaultNotificationPeriod: defaultNotificatinPeriod,
			}
			user, err = th.serv.CreateUser(ctx, user)
			if err != nil {
				handleError(ctx, b, chatID, err)

				return
			}
		}

		ctx = context.WithValue(ctx, userCtxKey, user)
		ctx = log.WithCtx(ctx, "userID", user.ID)

		next(ctx, b, update)
	}
}

type ctxKey string

const userCtxKey ctxKey = "userID"

var ErrNoUserInCtx = errors.New("no user id in context")

func UserFromCtx(ctx context.Context) (domain.User, error) {
	userID, ok := ctx.Value(userCtxKey).(domain.User)
	if !ok {
		return domain.User{}, ErrNoUserInCtx
	}

	return userID, nil
}

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc { //nolint:unused // for debug
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		log.Ctx(ctx).Debug("got update", "update", update)

		next(ctx, bot, update)
	}
}
