package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

type Notification struct {
	th             *TelegramHandler
	deletionCancel func()
}

func (th *TelegramHandler) Notify(ctx context.Context, notif domains.SendingNotification) error {
	op := "TelegramHandler.Notify: %w"
	user, err := th.userRepo.GetUserInfo(ctx, notif.Params.Params.Telegram)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	n := Notification{th: th, deletionCancel: nil}
	nd := notificationData{
		eventID:         notif.EventID,
		isNewStatusDone: true,
	}
	kb := inKbr.New(th.bot, inKbr.NoDeleteAfterClick()).Button("Done", nd.code(), errorHandling(n.setStatus))
	_, err = th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID:      notif.Params.Params.Telegram,
		Text:        notif.Message + " " + notif.NotificationTime.In(user.Location()).Format(dayTimeFormat),
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type notificationData struct {
	eventID         int
	isNewStatusDone bool
}

func (n *notificationData) code() []byte {
	return []byte(strconv.Itoa(n.eventID) + " " + strconv.FormatBool(n.isNewStatusDone))
}

func (n *notificationData) read(bts []byte) error {
	op := "notificationButtonData.read: %w"
	splitted := strings.Split(string(bts), " ")
	if len(splitted) != 2 { //nolint:gomnd // amount of parts in split
		return fmt.Errorf(op, fmt.Errorf("bad bts: %q", string(bts)))
	}

	eventID, err := strconv.Atoi(splitted[0])
	if err != nil {
		return fmt.Errorf(op, err)
	}

	newStatus, err := strconv.ParseBool(splitted[1])
	if err != nil {
		return fmt.Errorf(op, err)
	}
	n.eventID = eventID
	n.isNewStatusDone = newStatus

	return nil
}

const msgDeltionTime = 10 * time.Second

func (n *Notification) setStatus(ctx context.Context, b *bot.Bot, msg *models.Message, data []byte) error {
	op := "Notification.setStatus: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if n.deletionCancel != nil {
		n.deletionCancel()
	}

	var notifData notificationData
	err = notifData.read(data)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = n.th.serv.SetEventDoneStatus(ctx, user.ID, notifData.eventID, notifData.isNewStatusDone)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	notifData.isNewStatusDone = !notifData.isNewStatusDone
	text := "Event marked as done"
	btnText := "Cancel done"
	if notifData.isNewStatusDone {
		btnText = "Done"
		text = "Event undone"
	}
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Ok", nil, n.DeleteMsgInline).
		Row().Button(btnText, notifData.code(), errorHandling(n.setStatus))
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	deletionCtx, cancel := context.WithCancel(ctx)
	n.deletionCancel = cancel

	go func() {
		select {
		case <-deletionCtx.Done():
			log.Ctx(ctx).Info("select deletion ctx done")

			return
		case <-time.After(msgDeltionTime):
			log.Ctx(ctx).Info("select delete msg")
			n.DeleteMsgInline(ctx, b, msg, nil)
		}
	}()

	return nil
}

func (n *Notification) DeleteMsgInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	if n.deletionCancel != nil {
		n.deletionCancel()
	}
	log.Ctx(ctx).Info("delete msg")
	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		log.Ctx(ctx).Error("notify deletion", log.Err(err))
	}
}
