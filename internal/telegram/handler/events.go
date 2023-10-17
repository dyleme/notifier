package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/service"
)

var ErrCantParseMessage = errors.New("cant parse message")

func (th *TelegramHandler) EventsMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.EventsMenuInline: %w"

	listEvents := ListEvents{th: th}
	createEvents := NewEventCreation(th, true)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List events", nil, errorHandling(listEvents.listInline)).
		Row().Button("Create event", nil, onSelectErrorHandling(createEvents.SetTextMsg)).
		Row().Button("Create event from task", nil, errorHandling(createEvents.SelectTaskMsg)).
		Row().Button("Cancel", nil, errorHandling(th.MainMenuInline))

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Events actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type ListEvents struct {
	th *TelegramHandler
}

func (l *ListEvents) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "ListEvents.listInline: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	events, err := l.th.serv.ListEventsInPeriod(ctx, user.ID, time.Now(), time.Now().AddDate(1, 0, 0), service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if len(events) == 0 {
		kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
		kbr.Row().Button("Ok", nil, errorHandling(l.th.MainMenuInline))
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
			ChatID:    mes.Chat.ID,
			MessageID: mes.ID,
			Caption:   "No events",
		})

		if err != nil {
			return fmt.Errorf(op, err)
		}

		return nil
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, event := range events {
		ec := SingleEvent{th: l.th} //nolint:exhaustruct //fill it in ec.HandleBtnEventChosen
		text := event.Text + "\t|\t" + event.Start.In(user.Location()).Format(dayTimeFormat)
		kbr.Row().Button(text, []byte(strconv.Itoa(event.ID)), errorHandling(ec.HandleBtnEventChosen))
	}
	kbr.Row().Button("Cancel", nil, errorHandling(l.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "All events",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

const notSettedID = -1

func NewEventCreation(th *TelegramHandler, isWorkflow bool) SingleEvent {
	return SingleEvent{
		id:          notSettedID,
		th:          th,
		text:        "",
		description: "",
		date:        time.Time{},
		time:        time.Time{},
		isWorkflow:  isWorkflow,
	}
}

type SingleEvent struct {
	th          *TelegramHandler
	id          int
	text        string
	date        time.Time
	time        time.Time
	description string
	isWorkflow  bool
}

func (se *SingleEvent) next(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64,
	nextFunc func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error,
) error {
	op := "SingleEvent.next: %w"
	if se.isWorkflow {
		err := nextFunc(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	} else {
		err := se.EditMenuMsg(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	}

	return nil
}

func (se *SingleEvent) isCreation() bool {
	return se.id == notSettedID
}

func (se *SingleEvent) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !se.date.IsZero() {
		dateStr = se.date.Format(dayPointWithYearFormat)
	}
	if !se.time.IsZero() {
		timeStr = se.time.Format(timeDoublePointsFormat)
	}

	var eventStringBuilder strings.Builder
	eventStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", se.text))
	eventStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Description: %s\n", se.description))

	return eventStringBuilder.String()
}

func (se *SingleEvent) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set text", nil, onSelectErrorHandling(se.SetTextMsg)).
		Button("Set date", nil, onSelectErrorHandling(se.SetDateMsg)).
		Button("Set time", nil, onSelectErrorHandling(se.SetTimeMsg)).
		Button("Set description", nil, onSelectErrorHandling(se.SetDescription))

	kbr.Row()
	if se.isCreation() {
		kbr.Button("Create", nil, errorHandling(se.CreateInline))
	} else {
		kbr.Button("Update", nil, errorHandling(se.UpdateInline))
	}

	kbr.Button("Cancel", nil, errorHandling(se.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     se.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) SelectTaskMsg(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.MessageChooseTask: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	tasks, err := se.th.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, t := range tasks {
		kbr.Row().Button(t.Text, []byte(strconv.Itoa(t.ID)), errorHandling(se.HandleBtnTaskChosen))
	}

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     "Select task",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsTaskID []byte) error {
	op := "SingleEvent.HandleBtnTaskChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	taskID, err := strconv.Atoi(string(btsTaskID))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	task, err := se.th.serv.GetTask(ctx, taskID, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	se.text = task.Text

	err = se.next(ctx, b, msg.ID, msg.Chat.ID, se.SetDateMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter event text",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	se.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    se.HandleMsgSetText,
		messageID: relatedMsgID,
	})

	return nil
}

func (se *SingleEvent) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetText: %w"
	se.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	se.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = se.next(ctx, b, relatedMsgID, msg.Chat.ID, se.SetDateMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) SetDateMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetDateMsg: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	caption := se.String() + "\n\nEnter date (it can be or one of provided, or you can type your own date)"
	now := time.Now().In(user.Location())
	nowStr := now.Format(dayPointFormat)
	tomorrow := time.Now().Add(timeDay).In(user.Location())
	tomorrowStr := tomorrow.Format(dayPointFormat)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button(nowStr, []byte(nowStr), errorHandling(se.HandleBtnSetDate)).
		Row().Button(tomorrowStr, []byte(tomorrowStr), errorHandling(se.HandleBtnSetDate))

	se.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    se.HandleMsgSetDate,
		messageID: relatedMsgID,
	})

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     caption,
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleBtnSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error {
	op := "SingleEvent.HandleBtnSetDate: %w"

	if err := se.handleSetDate(ctx, b, msg.Chat.ID, msg.ID, string(bts)); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleMsgSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetDate: %w"

	if err := se.handleSetDate(ctx, b, msg.Chat.ID, relatedMsgID, msg.Text); err != nil {
		return fmt.Errorf(op, err)
	}

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) handleSetDate(ctx context.Context, b *bot.Bot, chatID int64, msgID int, dateStr string) error {
	op := "SingleEvent.handleSetDate: %w"

	t, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	se.date = t

	se.th.waitingActionsStore.Delete(chatID)

	err = se.next(ctx, b, msgID, chatID, se.SetTimeMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetTimeMsg: %w"
	caption := se.String() + "\n\nEnter time"

	se.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    se.HandleMsgSetTime,
		messageID: relatedMsgID,
	})
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   caption,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetTime: %w"

	t, err := parseTime(msg.Text)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	se.time = t
	se.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	se.isWorkflow = false
	err = se.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func computeTime(day, hourMinutes time.Time, loc *time.Location) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), hourMinutes.Hour(), hourMinutes.Minute(), 0, 0, loc)
}

func (se *SingleEvent) SetDescription(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetDescription: %w"
	caption := se.String() + "\n\nEnter description"

	se.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    se.HandleMsgSetDescription,
		messageID: relatedMsgID,
	})
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   caption,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleMsgSetDescription(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetDescription: %w"

	se.description = msg.Text
	se.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = se.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

var ErrTimeInPast = errors.New("time is in past")

func (se *SingleEvent) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	utcTime := computeTime(se.date, se.time, user.Location())
	if utcTime.Before(time.Now()) {
		return fmt.Errorf(op, ErrTimeInPast)
	}

	event := domains.Event{ //nolint:exhaustruct // don't know id on creation
		UserID:             user.ID,
		Text:               se.text,
		Description:        se.description,
		Start:              utcTime,
		Done:               false,
		Sended:             false,
		NotificationParams: nil,
		SendTime:           utcTime,
	}

	_, err = se.th.serv.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = se.th.MainMenuWithText(ctx, b, msg, "Service successfully created:\n"+se.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.UpdateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = se.th.serv.UpdateEvent(ctx, service.EventUpdateParams{
		ID:          se.id,
		Text:        se.text,
		UserID:      user.ID,
		Description: se.description,
		Start:       computeTime(se.date, se.time, user.Location()),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = se.th.MainMenuWithText(ctx, b, msg, "Service successfully updated:\n"+se.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = se.th.serv.DeleteEvent(ctx, user.ID, se.id)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = se.th.MainMenuWithText(ctx, b, msg, "Service successfully deleted:\n"+se.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (se *SingleEvent) HandleBtnEventChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsEventID []byte) error {
	op := "EventEdit.EventChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	eventID, err := strconv.Atoi(string(btsEventID))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	event, err := se.th.serv.GetEvent(ctx, user.ID, eventID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	se.id = event.ID
	se.date = event.Start
	se.time = event.Start
	se.text = event.Text
	se.description = event.Description
	se.isWorkflow = false

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, onSelectErrorHandling(se.EditMenuMsg)).
		Button("Delete", nil, errorHandling(se.DeleteInline)).
		Row().Button("Cancel", nil, errorHandling(se.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     se.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
