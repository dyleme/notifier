package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"
)

func (th *TelegramHandler) EventsMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	listEvents := ListEvents{th: th}
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List events", nil, errorHandling(listEvents.listInline)).
		Row().Button("Cancel", nil, errorHandling(th.MainMenuInline))

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Events actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("events menu inline: edit message caption: %w", err)
	}

	return nil
}

type ListEvents struct {
	th *TelegramHandler
}

func (le *ListEvents) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	events, err := le.th.serv.ListEvents(ctx, user.ID, timeborders.NewInfiniteUpper(time.Now()), defaultListParams)
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}

	if len(events) == 0 {
		kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
		kbr.Row().Button("Ok", nil, errorHandling(le.th.MainMenuInline))
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
			ChatID:    mes.Chat.ID,
			MessageID: mes.ID,
			Caption:   "No events",
		})
		if err != nil {
			return fmt.Errorf("edit message caption: %w", err)
		}

		return nil
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, event := range events {
		ev := Event{th: le.th} //nolint:exhaustruct //fill it in ev.HandleBtnTaskChosen
		text := event.Text + " " + event.SendTime.Format(dayTimeFormat)
		kbr.Row().Button(text, []byte(strconv.Itoa(event.ID)), errorHandling(ev.HandleBtnChosen))
	}
	kbr.Row().Button("Cancel", nil, errorHandling(le.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "All events",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("edit message caption: %w", err)
	}

	return nil
}

type Event struct {
	th         *TelegramHandler
	id         int
	text       string
	time       time.Time
	isWorkflow bool
}

func (ev *Event) HandleBtnChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsEventID []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	eventID, err := strconv.Atoi(string(btsEventID))
	if err != nil {
		return fmt.Errorf("strconv[string=%v]: %w", string(btsEventID), err)
	}

	event, err := ev.th.serv.GetEvent(ctx, eventID, user.ID)
	if err != nil {
		return fmt.Errorf("get periodic task[taskID=%v,userID=%v]: %w", eventID, user.ID, err)
	}

	ev.id = event.ID
	ev.text = event.Text
	ev.isWorkflow = false
	ev.time = event.SendTime

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, onSelectErrorHandling(ev.EditMenuMsg)).
		Button("Delete", nil, errorHandling(ev.DeleteInline)).
		Row().Button("Cancel", nil, errorHandling(ev.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     ev.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("edit message caption[chatID=%v,msgID=%v]: %w", msg.Chat.ID, msg.ID, err)
	}

	return nil
}

func (ev *Event) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set time", nil, onSelectErrorHandling(ev.SetTimeMsg))
	kbr.Row()
	kbr.Button("Update", nil, errorHandling(ev.UpdateInline))

	kbr.Button("Cancel", nil, errorHandling(ev.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     ev.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ev *Event) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !ev.time.IsZero() {
		dateStr = ev.time.Format(dayPointWithYearFormat)
		timeStr = ev.time.Format(timeDoublePointsFormat)
	}

	var taskStringBuilder strings.Builder
	taskStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", ev.text))
	taskStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	taskStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))

	return taskStringBuilder.String()
}

func (ev *Event) isCreation() bool {
	return ev.id == notSettedID
}

func (ev *Event) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	caption := ev.String() + "\n\nEnter time"

	ev.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    ev.HandleMsgSetTime,
		messageID: relatedMsgID,
	})
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   caption,
	})
	if err != nil {
		return fmt.Errorf("set time msg: edit message caption[chatID=%v,msgID=%v]: %w", chatID, relatedMsgID, err)
	}

	return nil
}

func (ev *Event) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	t, err := parseTime(msg.Text)
	if err != nil {
		return fmt.Errorf("hanle msg set time: parse time: %w", err)
	}
	ev.time = t
	ev.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("hanle msg set time: delete message: %w", err)
	}

	ev.isWorkflow = false
	err = ev.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("hanle msg set time: edit menu msg: %w", err)
	}

	return nil
}

func (ev *Event) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("update inline: user from ctx: %w", err)
	}

	err = ev.th.serv.ChangeEventTime(ctx, ev.id, ev.time, user.ID)
	if err != nil {
		return fmt.Errorf("change event time: %w", err)
	}

	err = ev.th.MainMenuWithText(ctx, b, msg, "Event successfully updated:\n"+ev.String())
	if err != nil {
		return fmt.Errorf("main menu with text: %w", err)
	}

	return nil
}

func (ev *Event) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("delete inline: user from ctx: %w", err)
	}

	err = ev.th.serv.DeleteEvent(ctx, user.ID, ev.id)
	if err != nil {
		return fmt.Errorf("delete inline: delete event: %w", err)
	}

	err = ev.th.MainMenuWithText(ctx, b, msg, "Service successfully deleted:\n"+ev.String())
	if err != nil {
		return fmt.Errorf("delete inline: main menu with text: %w", err)
	}

	return nil
}
