package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/domain"
	serverrors "github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/telegram/userinfo"
	"github.com/Dyleme/Notifier/pkg/log"
)

type Notification struct {
	th        *TelegramHandler
	done      bool
	id        int
	message   string
	notifTime time.Time
}

func (n *Notification) deleteOldNotificationMsg(ctx context.Context, eventID, chatID, newMsgID int) error {
	var oldMsgID int
	err := n.th.kvRepo.GetValue(ctx, strconv.Itoa(eventID), &oldMsgID)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			return fmt.Errorf("get message id [eventID=%v]: %w", eventID, err)
		}
	} else {
		log.Ctx(ctx).Info("got msgID", "eventID", eventID, "msgID", oldMsgID)
		_, err = n.th.bot.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: oldMsgID,
		})
		if err != nil {
			log.Ctx(ctx).Warn("can't delete msg", "msgID", oldMsgID, "chatID", chatID)
		}
	}

	log.Ctx(ctx).Info("save msgID", "eventID", eventID, "msgID", newMsgID)
	err = n.th.kvRepo.PutValue(ctx, strconv.Itoa(eventID), newMsgID)
	if err != nil {
		return fmt.Errorf("put message id [eventID=%v]: %w", eventID, err)
	}

	return nil
}

func (th *TelegramHandler) Notify(ctx context.Context, event domain.Notification) error {
	user, err := th.userRepo.GetUserInfo(ctx, event.Params.Params.Telegram)
	if err != nil {
		return fmt.Errorf("get user info[tgID=%v]: %w", event.Params.Params.Telegram, err)
	}
	n := Notification{
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

func (n *Notification) sendMessage(ctx context.Context, chatID int64, user userinfo.User) error {
	kb := inKbr.New(n.th.bot, inKbr.NoDeleteAfterClick()).
		Button("Done", nil, errorHandling(n.setDone)).
		Button("Reschedule", nil, onSelectErrorHandling(n.SetTimeMsg))
	text := n.message + " " + n.notifTime.In(user.Location()).Format(dayTimeFormat)
	msg, err := n.th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf("send message [chatID=%v, text=%q]: %w", chatID, text, err)
	}

	err = n.deleteOldNotificationMsg(ctx, n.id, int(chatID), msg.ID)
	if err != nil {
		log.Ctx(ctx).Error("update stored msg id", log.Err(err))
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

func (n *Notification) String() string {
	return n.message + "\n" + n.notifTime.Format(dayTimeFormat)
}

func (n *Notification) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	text := n.String() + "\n\nEnter time"

	n.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    n.HandleMsgSetTime,
		messageID: relatedMsgID,
	})
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Text:      text,
	})
	if err != nil {
		return fmt.Errorf("set time msg: edit message caption[chatID=%v,msgID=%v]: %w", chatID, relatedMsgID, err)
	}

	return nil
}

func (n *Notification) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("hanle msg set time: user from ctx: %w", err)
	}

	t, err := parseTime(msg.Text, user.Location())
	if err != nil {
		return fmt.Errorf("hanle msg set time: parse time: %w", err)
	}
	durFromDayStart := t.Sub(t.Truncate(timeDay))
	n.notifTime = n.notifTime.Truncate(timeDay).Add(durFromDayStart)
	n.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("hanle msg set time: delete message: %w", err)
	}

	err = n.SetDateMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("hanle msg set time: set date msg: %w", err)
	}

	return nil
}

func (n *Notification) SetDateMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetDateMsg: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	text := n.String() + "\n\nEnter date (it can bt or one of provided, or you can type your own date)"
	now := time.Now().In(user.Location())
	nowStr := now.Format(dayPointFormat)
	tomorrow := time.Now().Add(timeDay).In(user.Location())
	tomorrowStr := tomorrow.Format(dayPointFormat)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button(nowStr, []byte(nowStr), errorHandling(n.HandleBtnSetDate)).
		Row().Button(tomorrowStr, []byte(tomorrowStr), errorHandling(n.HandleBtnSetDate))

	n.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    n.HandleMsgSetDate,
		messageID: relatedMsgID,
	})

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Text:        text,
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (n *Notification) HandleBtnSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error {
	op := "SingleTask.HandleBtnSetDate: %w"

	if err := n.handleSetDate(ctx, b, msg.Chat.ID, msg.ID, string(bts)); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (n *Notification) HandleMsgSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	if err := n.handleSetDate(ctx, b, msg.Chat.ID, relatedMsgID, msg.Text); err != nil {
		return fmt.Errorf("handle msg set date: %w", err)
	}

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("handle msg set date: delete message: %w", err)
	}

	return nil
}

func (n *Notification) handleSetDate(ctx context.Context, b *bot.Bot, chatID int64, msgID int, dateStr string) error {
	t, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf("parse date: %w", err)
	}
	hmTime := n.notifTime.Sub(n.notifTime.Truncate(timeDay))
	n.notifTime = t.Add(hmTime)

	n.th.waitingActionsStore.Delete(chatID)

	err = n.Reschedule(ctx, b, msgID, chatID)
	if err != nil {
		return fmt.Errorf("edit menu msg: %w", err)
	}

	return nil
}

func (n *Notification) Reschedule(ctx context.Context, b *bot.Bot, msgID int, chatID int64) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	err = n.th.serv.ReschedulEventToTime(ctx, n.id, user.ID, n.notifTime)
	if err != nil {
		return fmt.Errorf("reschedule event: %w", err)
	}

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: msgID,
	})
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	return nil
}
