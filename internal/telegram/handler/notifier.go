package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"

	domains "github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func (th *TelegramHandler) Notify(ctx context.Context, notif domains.SendingNotification) error {
	op := "TelegramHandler.Notify: %w"
	kb := inline.New(th.bot.Bot).Button("Done", []byte(strconv.Itoa(notif.EventID)), th.markDone)
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      notif.Params.Params.Telegram,
		Text:        notif.Message,
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) markDone(ctx context.Context, _ *bot.Bot, mes *models.Message, data []byte) {
	op := "TelegramHandler.markDone: %w"
	chatID := mes.Chat.ID
	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}
	eventID, err := strconv.Atoi(string(data))
	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}
	_, err = th.serv.UpdateEvent(ctx, service.UpdateEventParams{
		ID:     eventID,
		UserID: userID,
		Done:   true,
	})

	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}

	_, err = th.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: mes.Chat.ID,
		Text:   "Event is done",
	})
	if err != nil {
		th.handleError(ctx, chatID, fmt.Errorf(op, err))

		return
	}
}
