package handler

import (
	"context"
	"fmt"

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
	op := "TelegramHandler.info: %w"
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "This is info from schedulder bot",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
