package handler

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"
)

func (th *TelegramHandler) SendBindingMessage(ctx context.Context, tgID int, code string) error {
	kb := inKbr.New(th.bot).
		Button("DeleteMessage", nil, func(_ context.Context, _ *bot.Bot, _ models.MaybeInaccessibleMessage, _ []byte) {})
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct // no need to fill
		ChatID:      tgID,
		Text:        "Your code: " + code,
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
