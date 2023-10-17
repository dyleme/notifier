package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
)

type Notification struct {
	th        *TelegramHandler
	eventType domains.EventType
	done      bool
	eventID   int
	message   string
	notifTime time.Time
}

func (th *TelegramHandler) Notify(ctx context.Context, notif domains.SendingNotification) error {
	user, err := th.userRepo.GetUserInfo(ctx, notif.Params.Params.Telegram)
	if err != nil {
		return fmt.Errorf("get user info[tgID=%v]: %w", notif.Params.Params.Telegram, err)
	}
	n := Notification{
		th:        th,
		eventType: notif.EventType,
		done:      false,
		eventID:   notif.EventID,
		message:   notif.Message,
		notifTime: notif.NotificationTime,
	}
	err = n.sendMessage(ctx, int64(notif.Params.Params.Telegram), user)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (n *Notification) sendMessage(ctx context.Context, chatID int64, user userinfo.User) error {
	kb := inKbr.New(n.th.bot, inKbr.NoDeleteAfterClick()).Button("Done", nil, errorHandling(n.setDone))
	_, err := n.th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID:      chatID,
		Text:        n.message + " " + n.notifTime.In(user.Location()).Format(dayTimeFormat),
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (n *Notification) setUndone(ctx context.Context, _ *bot.Bot, msg *models.Message, _ []byte) error {
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

func (n *Notification) setDone(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
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

func (n *Notification) SendDone(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	err = n.th.serv.SetEventDoneStatus(ctx, service.AbstractEvent{
		EventID:   n.eventID,
		EventType: n.eventType,
		UserID:    user.ID,
		Done:      true,
	})
	if err != nil {
		return fmt.Errorf("set event done status: %w", err)
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
