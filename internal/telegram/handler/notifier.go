package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
)

type Event struct {
	th        *TelegramHandler
	done      bool
	id        int
	message   string
	notifTime time.Time
}

func (th *TelegramHandler) Notify(ctx context.Context, event domains.SendingEvent) error {
	user, err := th.userRepo.GetUserInfo(ctx, event.Params.Params.Telegram)
	if err != nil {
		return fmt.Errorf("get user info[tgID=%v]: %w", event.Params.Params.Telegram, err)
	}
	n := Event{
		th:        th,
		done:      false,
		id:        event.EventID,
		message:   event.Message,
		notifTime: event.SendTime,
	}
	err = n.sendMessage(ctx, int64(user.TGID), user)
	if err != nil {
		return fmt.Errorf("send message [user: %v]: %w", user, err)
	}

	return nil
}

func (n *Event) sendMessage(ctx context.Context, chatID int64, user userinfo.User) error {
	kb := inKbr.New(n.th.bot, inKbr.NoDeleteAfterClick()).Button("Done", nil, errorHandling(n.setDone))
	text := n.message + " " + n.notifTime.In(user.Location()).Format(dayTimeFormat)
	_, err := n.th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf("send message [chatID=%v, text=%q]: %w", chatID, text, err)
	}

	return nil
}

func (n *Event) setUndone(ctx context.Context, _ *bot.Bot, msg *models.Message, _ []byte) error {
	n.done = false
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	err = n.sendMessage(ctx, msg.Chat.ID, user)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (n *Event) setDone(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	n.done = true

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Yes", nil, errorHandling(n.SendDone)).
		Row().Button("No", nil, errorHandling(n.setUndone))
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        "Are you sure?",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("edit message text: %w", err)
	}

	return nil
}

func (n *Event) SendDone(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	err = n.th.serv.SetEventDoneStatus(ctx, n.id, user.ID, n.done)
	if err != nil {
		return fmt.Errorf("set task done status [eventID=%v, userID=%v]: %w", n.id, user.ID, err)
	}

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	return nil
}
