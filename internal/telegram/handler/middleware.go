package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
)

func (th *TelegramHandler) tgUserID(update *models.Update) (int64, error) {
	switch {
	case update.Message != nil:
		return update.Message.From.ID, nil
	case update.CallbackQuery != nil:
		return update.CallbackQuery.Sender.ID, nil
	}

	return 0, errors.New("unknown id")
}

func (th *TelegramHandler) UserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatID, err := th.chatID(update)
		if err != nil {
			handleError(ctx, b, 0, err)

			return
		}

		tgUserID, err := th.tgUserID(update)
		if err != nil {
			handleError(ctx, b, chatID, err)

			return
		}

		userInfo, err := th.userRepo.GetUserInfo(ctx, int(tgUserID))
		if err != nil {
			handleError(ctx, b, chatID, err)

			return
		}

		ctx = context.WithValue(ctx, userCtxKey, userInfo)

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

func loggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc { //nolint:unused // no need for now
	return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
		log.Ctx(ctx).Debug("got update", "update", update)

		next(ctx, bot, update)
	}
}

func chatID(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.EditedMessage != nil {
		return update.EditedMessage.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Sender.ID
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
