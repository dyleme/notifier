package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

var ErrCantParseMessage = errors.New("cant parse message")

func (th *TelegramHandler) EventsMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) {
	listEvents := ListEvents{th: th}
	createEvents := NewEventCreation(th, true)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List events", nil, listEvents.listInline).
		Row().Button("Create event", nil, makeOnSelect(createEvents.SetTextMsg)).
		Row().Button("Create event from task", nil, createEvents.SelectTaskMsg)

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Events actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, mes.Chat.ID, err)
	}
}

type ListEvents struct {
	th *TelegramHandler
}

func (l *ListEvents) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) {
	op := "ListEvents.listInline: %w"

	userInfo, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, mes.Chat.ID, fmt.Errorf(op, err))

		return
	}

	events, err := l.th.serv.ListEvents(ctx, userInfo.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		handleError(ctx, b, mes.Chat.ID, fmt.Errorf(op, err))

		return
	}

	if len(events) == 0 {
		kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
		kbr.Row().Button("Ok", nil, l.th.MainMenuInline)
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
			ChatID:    mes.Chat.ID,
			MessageID: mes.ID,
			Caption:   "No events",
		})

		if err != nil {
			handleError(ctx, b, mes.Chat.ID, fmt.Errorf(op, err))

			return
		}

		return
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, event := range events {
		ec := EventCreation{th: l.th} //nolint:exhaustruct //fill it in ec.HandleBtnEventChosen
		kbr.Row().Button(event.Text, []byte(strconv.Itoa(event.ID)), ec.HandleBtnEventChosen)
	}

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "All events",
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, mes.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

const notSettedID = -1

func NewEventCreation(th *TelegramHandler, isWorkflow bool) EventCreation {
	return EventCreation{
		id:          notSettedID,
		th:          th,
		text:        "",
		description: "",
		day:         time.Time{},
		time:        time.Time{},
		isWorkflow:  isWorkflow,
	}
}

type EventCreation struct {
	th          *TelegramHandler
	id          int
	text        string
	day         time.Time
	time        time.Time
	description string
	isWorkflow  bool
}

func (ec *EventCreation) next(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64,
	nextF func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64),
) {
	if ec.isWorkflow {
		nextF(ctx, b, relatedMsgID, chatID)
	} else {
		ec.EditMenuMsg(ctx, b, relatedMsgID, chatID)
	}
}

func (ec *EventCreation) isCreation() bool {
	return ec.id == notSettedID
}

func (ec *EventCreation) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !ec.day.IsZero() {
		dateStr = ec.day.Format(dayPointWithYearFormat)
	}
	if !ec.time.IsZero() {
		timeStr = ec.time.Format(timeDoublePointsFormat)
	}

	var eventStringBuilder strings.Builder
	eventStringBuilder.WriteString("Current event\n")
	eventStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", ec.text))
	eventStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Description: %s\n", ec.description))

	return eventStringBuilder.String()
}

func (ec *EventCreation) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) {
	op := "EventCreation.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Change text", nil, makeOnSelect(ec.SetTextMsg)).
		Button("Change date", nil, makeOnSelect(ec.SetDateMsg)).
		Button("Change time", nil, makeOnSelect(ec.SetTimeMsg))

	kbr.Row()
	if ec.isCreation() {
		kbr.Button("Create", nil, ec.CreateInline)
	} else {
		kbr.Button("Update", nil, ec.UpdateInline)
	}

	kbr.Button("Cancel", nil, ec.th.MainMenuInline)

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     ec.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))
	}
}

func (ec *EventCreation) SelectTaskMsg(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "EventCreation.MessageChooseTask: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	tasks, err := ec.th.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, t := range tasks {
		kbr.Row().Button(t.Text, []byte(strconv.Itoa(t.ID)), ec.HandleBtnTaskChosen)
	}

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     "Select task",
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

func (ec *EventCreation) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsTaskID []byte) {
	op := "EventCreation.HandleBtnTaskChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	taskID, err := strconv.Atoi(string(btsTaskID))
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	task, err := ec.th.serv.GetTask(ctx, taskID, user.ID)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	ec.text = task.Text

	ec.next(ctx, b, msg.ID, msg.Chat.ID, ec.SetDateMsg)
}

func (ec *EventCreation) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) {
	op := "EventCreation.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter event text",
	})
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))
	}

	ec.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    ec.HandleMsgSetText,
		messageID: relatedMsgID,
	})
}

func (ec *EventCreation) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) {
	op := "EventCreation.HandleMsgSetText: %w"
	ec.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	ec.th.waitingActionsStore.Delete(msg.Chat.ID)

	ec.next(ctx, b, relatedMsgID, msg.Chat.ID, ec.SetDateMsg)
}

func (ec *EventCreation) SetDateMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) {
	op := "EventCreation.SetDateMsg: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}
	caption := ec.String() + "\n\nEnter date (it can be or one of provided, or you can type your own date)"
	now := time.Now().In(user.Location())
	nowBts, err := json.Marshal(now)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}
	tomorrow := time.Now().Add(day).In(user.Location())
	tomorrowBts, err := json.Marshal(tomorrow)
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))

		return
	}
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button(now.Format(dayPointFormat), nowBts, ec.HandleBtnSetDate).
		Row().Button(tomorrow.Format(dayPointFormat), tomorrowBts, ec.HandleBtnSetDate)

	ec.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    ec.HandleMsgSetDate,
		messageID: relatedMsgID,
	})

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     caption,
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))
	}
}

func (ec *EventCreation) HandleBtnSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) {
	op := "EventCreation.HandleBtnSetDate: %w"

	var t time.Time
	err := json.Unmarshal(bts, &t)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
	ec.day = t
	ec.th.waitingActionsStore.Delete(msg.Chat.ID)

	ec.next(ctx, b, msg.ID, msg.Chat.ID, ec.SetTimeMsg)
}

func (ec *EventCreation) HandleMsgSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) {
	op := "EventCreation.HandleMsgSetDate: %w"

	t, err := parseDay(msg.Text)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
	ec.day = t
	ec.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	ec.next(ctx, b, relatedMsgID, msg.Chat.ID, ec.SetTimeMsg)
}

func (ec *EventCreation) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) {
	op := "EventCreation.SetTimeMsg: %w"
	caption := ec.String() + "\n\nEnter time"

	ec.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    ec.HandleMsgSetTime,
		messageID: relatedMsgID,
	})
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   caption,
	})
	if err != nil {
		handleError(ctx, b, chatID, fmt.Errorf(op, err))
	}
}

func (ec *EventCreation) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) {
	op := "EventCreation.HandleMsgSetTime: %w"

	t, err := parseTime(msg.Text)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
	ec.time = t
	ec.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	ec.isWorkflow = false
	ec.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
}

func computeTime(day, hourMinutes time.Time, loc *time.Location) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), hourMinutes.Hour(), hourMinutes.Minute(), 0, 0, loc)
}

func (ec *EventCreation) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "EventCreation.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	event := domains.Event{ //nolint:exhaustruct // don't know id on creation
		UserID:      user.ID,
		Text:        ec.text,
		Description: "",
		Start:       computeTime(ec.day, ec.time, user.Location()),
		Done:        false,
		Notification: domains.Notification{
			Sended:             false,
			NotificationParams: nil,
		},
	}

	_, err = ec.th.serv.CreateEvent(ctx, event)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: msg.Chat.ID,
		Text:   "Event successfully created",
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

func (ec *EventCreation) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "EventCreation.UpdateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))
	}

	_, err = ec.th.serv.UpdateEvent(ctx, service.EventUpdateParams{
		ID:          ec.id,
		Text:        ec.text,
		UserID:      user.ID,
		Description: "",
		Start:       computeTime(ec.day, ec.time, user.Location()),
		Done:        false,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: msg.Chat.ID,
		Text:   "Task successfully updated",
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

func (ec *EventCreation) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "EventCreation.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	err = ec.th.serv.DeleteEvent(ctx, user.ID, ec.id)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to fill
		ChatID: msg.Chat.ID,
		Text:   "Deleted successfully",
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}

func (ec *EventCreation) HandleBtnEventChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsEventID []byte) {
	op := "EventEdit.EventChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	eventID, err := strconv.Atoi(string(btsEventID))
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	event, err := ec.th.serv.GetEvent(ctx, user.ID, eventID)
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}

	ec.id = event.ID
	ec.day = event.Start
	ec.time = event.Start
	ec.text = event.Text
	ec.description = event.Description
	ec.isWorkflow = false

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, makeOnSelect(ec.EditMenuMsg)).
		Button("Delete", nil, ec.DeleteInline).
		Row().Button("Cancel", nil, ec.th.MainMenuInline)

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     ec.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))

		return
	}
}
