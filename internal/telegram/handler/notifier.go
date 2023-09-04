package handler

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"

	domains "github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (th *TelegramHandler) Notify(ctx context.Context, notif domains.SendingNotification) error {
	op := "TelegramHandler.Notify: %w"
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: notif.Params.Params.Telegram,
		Text:   notif.Message,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
