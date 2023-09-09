package handler

import (
	"context"

	"github.com/go-telegram/bot"

	"github.com/Dyleme/Notifier/internal/lib/tgwf"
)

func (th *TelegramHandler) ShowInfo(ctx context.Context, _ *bot.Bot, chatID int64) (tgwf.Handler, error) {
	err := th.info(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (th *TelegramHandler) info(ctx context.Context, chatID int64) error {
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "info",
	})
	if err != nil {
		return err
	}

	return nil
}
