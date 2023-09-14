package tgwf

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dyleme/timecache"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

type WorkflowHandler struct {
	*bot.Bot
	lastActions   *timecache.Cache[int64, Handler]
	defaultAction Action
}

func New(b *bot.Bot, defaultAction Action) *WorkflowHandler {
	return &WorkflowHandler{
		Bot:           b,
		lastActions:   timecache.New[int64, Handler](),
		defaultAction: defaultAction,
	}
}

var ErrNoAssociatedAction = errors.New("there is no assicatiated action")

func (wh *WorkflowHandler) HandleAction(ctx context.Context, update *models.Update) error {
	chatID, err := wh.chatID(update)
	if err != nil {
		return err
	}

	handler, ok := wh.currentHandler(ctx, chatID)
	if !ok {
		return ErrNoAssociatedAction
	}

	err = wh.handleAction(ctx, handler, update, chatID)
	if err != nil {
		return err
	}

	return nil
}

func (wh *WorkflowHandler) Start(ctx context.Context, chatID int64, action Action) error {
	handler, err := action(ctx, wh.Bot, chatID)
	if err != nil {
		return err
	}

	if handler == nil {
		if wh.defaultAction != nil {
			err = wh.Start(ctx, chatID, wh.defaultAction)
			if err != nil {
				return err
			}
		}

		return nil
	}

	wh.lastActions.StoreDefDur(chatID, handler)
	log.Ctx(ctx).Debug("store", "type", fmt.Sprintf("%T", handler))

	return nil
}

func (wh *WorkflowHandler) handleAction(ctx context.Context, handler Handler, update *models.Update, chatID int64) error {
	nextAction, err := handler(ctx, wh.Bot, update)
	if err != nil {
		return err
	}
	if nextAction == nil {
		if wh.defaultAction == nil {
			wh.lastActions.Delete(chatID)
			log.Ctx(ctx).Debug("delete", "chatID", chatID)

			return nil
		}

		nextAction = wh.defaultAction
	}

	err = wh.Start(ctx, chatID, nextAction)
	if err != nil {
		return err
	}

	return nil
}

func (wh *WorkflowHandler) ForgotForChat(_ context.Context, chatID int64) {
	wh.lastActions.Delete(chatID)
}

func (wh *WorkflowHandler) currentHandler(ctx context.Context, userID int64) (Handler, bool) {
	handler, err := wh.lastActions.Get(userID)
	log.Ctx(ctx).Debug("get", "type", fmt.Sprintf("%T", handler))
	if err != nil {
		return nil, false
	}

	return handler, true
}

func (wh *WorkflowHandler) chatID(update *models.Update) (int64, error) {
	switch {
	case update.Message != nil:
		return update.Message.Chat.ID, nil
	case update.CallbackQuery != nil:
		if update.CallbackQuery.Message != nil {
			return update.CallbackQuery.Message.Chat.ID, nil
		}
	}

	return 0, fmt.Errorf("no chat id")
}
