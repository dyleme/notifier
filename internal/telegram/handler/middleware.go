package handler

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

func (th *TelegramHandler) UserIDMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
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

		userID, err := th.userRepo.GetID(ctx, int(tgUserID))
		if err != nil {
			th.handleError(ctx, chatID, err)

			return
		}

		ctx = context.WithValue(ctx, userIDCtxKey, userID)

		next(ctx, bot, update)
	}
}

type ctxKey string

const userIDCtxKey ctxKey = "userID"

var ErrNoUserIDInCtx = errors.New("no user id in context")

func UserIDFromCtx(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(userIDCtxKey).(int)
	if !ok {
		return 0, ErrNoUserIDInCtx
	}

	return userID, nil
}

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		log.Ctx(ctx).Debug("got update", "update", update)

		next(ctx, bot, update)
	}
}
