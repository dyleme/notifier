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
	Fields [][]MenuActionField
}

func NewMenuAction(text string) *MenuAction {
	return &MenuAction{Text: text, hide: true}
}

func (ma *MenuAction) Btn(text string, action Action) *MenuAction {
	if len(ma.Fields) == 0 {
		ma.Fields = append(ma.Fields, []MenuActionField{})
	}
	ma.Fields[len(ma.Fields)-1] = append(ma.Fields[len(ma.Fields)-1], MenuActionField{
		Text:       text,
		NextAction: action,
	})

	return ma
}

func (ma *MenuAction) Row() *MenuAction {
	ma.Fields = append(ma.Fields, []MenuActionField{})
	return ma
}

func AddSliceToMenu[T any](ma *MenuAction, ts []T, btnText func(t T) string, action Action) *MenuAction {
	for i, t := range ts {
		text := strconv.Itoa(i) + "." + btnText(t)
		ma = ma.Row().Btn(text, action)
	}
	return ma
}

func (ma *MenuAction) SetHide(hide bool) *MenuAction {
	ma.hide = hide
	return ma
}

func (ma *MenuAction) Show(ctx context.Context, b *bot.Bot, chatID int64) (Handler, error) {
	keyboard := make([][]models.KeyboardButton, 0, len(ma.Fields))
	for _, row := range ma.Fields {
		keysRow := make([]models.KeyboardButton, 0, len(row))
		for _, btn := range row {
			keysRow = append(keysRow, models.KeyboardButton{Text: btn.Text})
		}
		keyboard = append(keyboard, keysRow)
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   ma.Text,
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			OneTimeKeyboard: ma.hide,
		},
	})
	if err != nil {
		return nil, err
	}

	return ma.Post, nil
}

func (ma *MenuAction) Post(ctx context.Context, b *bot.Bot, update *models.Update) (Action, error) {
	message, err := GetMessage(update)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ma.Fields); i++ {
		for j := 0; j < len(ma.Fields[i]); j++ {
			if ma.Fields[i][j].Text == message.Text {
				return ma.Fields[i][j].NextAction, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown message")
}
