package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
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
				handleError(log.WithCtx(ctx, "panic", "true"), bot, chatID(update), fmt.Errorf("%v", r))
			}
		}()

		next(ctx, bot, update)
	}
}

func (th *TelegramHandler) UserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID, err := th.chatID(update)
		if err != nil {
			handleError(ctx, b, 0, err)

			return
		}

		var tgUserID int64
		var nickname string
		switch {
		case update.Message != nil:
			tgUserID = update.Message.From.ID
			nickname = update.Message.From.Username
		case update.CallbackQuery != nil:
			tgUserID = update.CallbackQuery.From.ID
			nickname = update.CallbackQuery.From.Username
		}

		var userInfo userinfo.User
		userInfo, err = th.userRepo.GetUserInfo(ctx, int(tgUserID))
		if err != nil {
			var notFoundErr serverrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				handleError(ctx, b, chatID, err)

				return
			}

			userInfo, err = th.userRepo.AddUser(ctx, int(tgUserID), nickname)
			if err != nil {
				handleError(ctx, b, chatID, err)

				return
			}
		}

		ctx = context.WithValue(ctx, userCtxKey, userInfo)
		log.WithCtx(ctx, "userID", userInfo.ID)

		next(ctx, b, update)
	}
}

type ctxKey string

const userCtxKey ctxKey = "userID"

var ErrNoUserInCtx = errors.New("no user id in context")

func UserFromCtx(ctx context.Context) (userinfo.User, error) {
	userID, ok := ctx.Value(userCtxKey).(userinfo.User)
	if !ok {
		return userinfo.User{}, ErrNoUserInCtx
	}

	return userID, nil
}

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		log.Ctx(ctx).Debug("got update", "update", update)

		next(ctx, bot, update)
	}
}
