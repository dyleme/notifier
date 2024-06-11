package handler

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	_ "embed"
)

func (th *TelegramHandler) mainMenuKeyboard(b *bot.Bot) models.ReplyMarkup {
	return inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Info", nil, errorHandling(th.InfoInline)).
		Row().Button("Tasks", nil, errorHandling(th.TasksMenuInline)).
		Row().Button("Periodic tasks", nil, errorHandling(th.PeriodicTasksMenuInline)).
		Row().Button("Settings", nil, errorHandling(th.SettingsInline))
}

const imageName = "clock.png"

//go:embed clock.png
var image []byte

func (th *TelegramHandler) mainMenuCreateWindow(ctx context.Context, b *bot.Bot, chatID int64) error {
	op := "TelegramHandler.mainMenuCreateWindow: %w"

	err := th.SendImage(ctx, imageName, image, &bot.SendPhotoParams{ //nolint:exhaustruct //no need to fill
		Caption:     "Main menu",
		ChatID:      chatID,
		ReplyMarkup: th.mainMenuKeyboard(b),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) AdditionalText(inlineFunc func(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte, text string) error, text string) func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TelegramHandler.AdditionalText: %w"

	return func(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error {
		if err := inlineFunc(ctx, b, msg, bts, text); err != nil {
			return fmt.Errorf(op, err)
		}

		return nil
	}
}

func (th *TelegramHandler) MainMenuWithText(ctx context.Context, b *bot.Bot, msg *models.Message, text string) error {
	op := "TelegramHandler.MainMenuWithText: %w"
	caption := "Main menu"
	if text != "" {
		caption = caption + "\n\n" + text
	}

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct // no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     caption,
		ReplyMarkup: th.mainMenuKeyboard(b),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) MainMenuInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TelegramHandler.MainMenuInline: %w"
	err := th.MainMenuWithText(ctx, b, msg, "")
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) InfoInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.InfoInline: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Main menu", nil, errorHandling(th.MainMenuInline))

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Edit caption",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
