package handler

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"
)

func (th *TelegramHandler) mainMenuKeyboard(b *bot.Bot) models.ReplyMarkup {
	return inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Info", nil, th.InfoInline).
		Row().Button("Events", nil, th.EventsMenuInline)
}

const imageName = "clock.png"

//go:embed clock.png
var image []byte

func (th *TelegramHandler) mainMenuCreateWindow(ctx context.Context, b *bot.Bot, chatID int64) error {
	op := "TelegramHandler.mainMenuCreateWindow: %w"

	err := th.SendImage(ctx, imageName, image, &bot.SendPhotoParams{ //nolint:exhaustruct //no need to fill
		Caption:     "EventChosen",
		ChatID:      chatID,
		ReplyMarkup: th.mainMenuKeyboard(b),
	})
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return err
	}

	return nil
}

func (th *TelegramHandler) MainMenuInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "TelegramHandler.MainMenuInline: %w"

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct // no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     "Edit caption",
		ReplyMarkup: th.mainMenuKeyboard(b),
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

func (th *TelegramHandler) InfoInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) {
	op := "TelegramHandler.InfoInline: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Main menu", nil, th.MainMenuInline)

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Edit caption",
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, mes.Chat.ID, fmt.Errorf(op, err))
	}
}
