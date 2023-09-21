package tgwf

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MenuActionField struct {
	Text       string
	NextAction Action
}

type MenuAction struct {
	Text   string
	hide   bool
	fields [][]MenuActionField
}

func NewMenuAction(text string) *MenuAction {
	return &MenuAction{Text: text, hide: true, fields: nil}
}

func (ma *MenuAction) Btn(text string, action Action) *MenuAction {
	if len(ma.fields) == 0 {
		ma.fields = append(ma.fields, []MenuActionField{})
	}
	ma.fields[len(ma.fields)-1] = append(ma.fields[len(ma.fields)-1], MenuActionField{
		Text:       text,
		NextAction: action,
	})

	return ma
}

func (ma *MenuAction) Row() *MenuAction {
	ma.fields = append(ma.fields, []MenuActionField{})

	return ma
}

func AddSliceToMenu[T any](ma *MenuAction, ts []T, btnText func(t T) string, action Action) *MenuAction {
	for i, t := range ts {
		text := strconv.Itoa(i+1) + "." + btnText(t)
		ma = ma.Row().Btn(text, action)
	}

	return ma
}

func (ma *MenuAction) SetHide(hide bool) *MenuAction {
	ma.hide = hide

	return ma
}

func (ma *MenuAction) Show(ctx context.Context, b *bot.Bot, chatID int64) (Handler, error) {
	op := "MenuAction.Show: %w"
	keyboard := make([][]models.KeyboardButton, 0, len(ma.fields))
	for _, row := range ma.fields {
		keysRow := make([]models.KeyboardButton, 0, len(row))
		for _, btn := range row {
			keysRow = append(keysRow, models.KeyboardButton{Text: btn.Text}) //nolint:exhaustruct //no need to specify
		}
		keyboard = append(keyboard, keysRow)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   ma.Text,
		ReplyMarkup: models.ReplyKeyboardMarkup{ //nolint:exhaustruct //no need to specify
			Keyboard:        keyboard,
			OneTimeKeyboard: ma.hide,
		},
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ma.Post, nil
}

func (ma *MenuAction) Post(_ context.Context, _ *bot.Bot, update *models.Update) (Action, error) {
	message, err := GetMessage(update)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ma.fields); i++ {
		for j := 0; j < len(ma.fields[i]); j++ {
			if ma.fields[i][j].Text == message.Text {
				return ma.fields[i][j].NextAction, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown message")
}
