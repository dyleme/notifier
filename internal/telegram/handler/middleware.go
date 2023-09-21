package handler

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
)

func (th *TelegramHandler) UserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		chatID, err := th.chatID(update)
		if err != nil {
			th.handleError(ctx, 0, err)

			return
		}

		tgUserID, err := th.tgUserID(update)
		if err != nil {
			th.handleError(ctx, chatID, err)

			return
		}

		userInfo, err := th.userRepo.GetUserInfo(ctx, int(tgUserID))
		if err != nil {
			th.handleError(ctx, chatID, err)

			return
		}

		ctx = context.WithValue(ctx, userCtxKey, userInfo)

		next(ctx, bot, update)
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
