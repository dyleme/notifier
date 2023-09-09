package tgwf

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type (
	Action  func(ctx context.Context, b *bot.Bot, chatID int64) (Handler, error)
	Handler func(ctx context.Context, b *bot.Bot, update *models.Update) (Action, error)
)
