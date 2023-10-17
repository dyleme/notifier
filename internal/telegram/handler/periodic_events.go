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

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/service"
)

func (th *TelegramHandler) PeriodicEventsMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.EventsMenuInline: %w"

	listEvents := ListPeriodicEvents{th: th}
	createEvents := NewPeriodicEventCreation(th, true)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List periodic events", nil, errorHandling(listEvents.listInline)).
		Row().Button("Create periodic event", nil, onSelectErrorHandling(createEvents.SetTextMsg)).
		Row().Button("Create periodic event from task", nil, errorHandling(createEvents.SelectTaskMsg)).
		Row().Button("Cancel", nil, errorHandling(th.MainMenuInline))

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Periodic events actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type ListPeriodicEvents struct {
	th *TelegramHandler
}

func (l *ListPeriodicEvents) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "ListEvents.listInline: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	events, err := l.th.serv.ListFuturePeriodicEvents(ctx, user.ID, defaultListParams)
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
		pe := PeriodicEvent{th: l.th} //nolint:exhaustruct //fill it in pe.HandleBtnEventChosen
		text := event.Text + "\t|\t" + event.Notification.SendTime.In(user.Location()).Format(dayTimeFormat)
		kbr.Row().Button(text, []byte(strconv.Itoa(event.ID)), errorHandling(pe.HandleBtnEventChosen))
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

func NewPeriodicEventCreation(th *TelegramHandler, isWorkflow bool) PeriodicEvent {
	return PeriodicEvent{
		id:             notSettedID,
		th:             th,
		text:           "",
		description:    "",
		smallestPeriod: 0,
		biggestPeriod:  0,
		time:           time.Time{},
		isWorkflow:     isWorkflow,
	}
}

type PeriodicEvent struct {
	th             *TelegramHandler
	id             int
	text           string
	smallestPeriod time.Duration
	biggestPeriod  time.Duration
	time           time.Time
	description    string
	isWorkflow     bool
}

func (pe *PeriodicEvent) next(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64,
	nextFunc func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error,
) error {
	op := "SingleEvent.next: %w"
	if pe.isWorkflow {
		err := nextFunc(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	} else {
		err := pe.EditMenuMsg(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	}

	return nil
}

func (pe *PeriodicEvent) isCreation() bool {
	return pe.id == notSettedID
}

func durToString(dur time.Duration) string {
	days := dur / timeDay
	if days == 0 {
		return ""
	}

	return strconv.Itoa(int(days))
}

func (pe *PeriodicEvent) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !pe.time.IsZero() {
		timeStr = pe.time.Format(timeDoublePointsFormat)
	}

	var eventStringBuilder strings.Builder
	eventStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", pe.text))
	eventStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))
	eventStringBuilder.WriteString(fmt.Sprintf("Smallest period: %v\n", durToString(pe.smallestPeriod)))
	eventStringBuilder.WriteString(fmt.Sprintf("Biggest period: %v\n", durToString(pe.biggestPeriod)))
	eventStringBuilder.WriteString(fmt.Sprintf("Description: %s\n", pe.description))

	return eventStringBuilder.String()
}

func (pe *PeriodicEvent) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set text", nil, onSelectErrorHandling(pe.SetTextMsg)).
		Button("Set smallest period", nil, onSelectErrorHandling(pe.SetSmallestPeriodMsg)).
		Button("Set biggest period", nil, onSelectErrorHandling(pe.SetBiggestPeriodMsg)).
		Button("Set time", nil, onSelectErrorHandling(pe.SetTimeMsg)).
		Button("Set description", nil, onSelectErrorHandling(pe.SetDescription))

	kbr.Row()
	if pe.isCreation() {
		kbr.Button("Create", nil, errorHandling(pe.CreateInline))
	} else {
		kbr.Button("Update", nil, errorHandling(pe.UpdateInline))
	}

	kbr.Button("Cancel", nil, errorHandling(pe.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     pe.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) SelectTaskMsg(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.MessageChooseTask: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	tasks, err := pe.th.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, t := range tasks {
		kbr.Row().Button(t.Text, []byte(strconv.Itoa(t.ID)), errorHandling(pe.HandleBtnTaskChosen))
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

func (pe *PeriodicEvent) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsTaskID []byte) error {
	op := "SingleEvent.HandleBtnTaskChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	taskID, err := strconv.Atoi(string(btsTaskID))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	task, err := pe.th.serv.GetTask(ctx, taskID, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pe.text = task.Text

	err = pe.next(ctx, b, msg.ID, msg.Chat.ID, pe.SetSmallestPeriodMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter event text",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pe.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pe.HandleMsgSetText,
		messageID: relatedMsgID,
	})

	return nil
}

func (pe *PeriodicEvent) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetText: %w"
	pe.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pe.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pe.next(ctx, b, relatedMsgID, msg.Chat.ID, pe.SetSmallestPeriodMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) SetSmallestPeriodMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "PeriodicEvent.SetSmallestPeriodMsg: %w"
	caption := pe.String() + "\n\nEnter smallest amount of days in period"
	pe.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pe.HandleMsgSetSmallestPeriod,
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

func (pe *PeriodicEvent) HandleMsgSetSmallestPeriod(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	days, err := strconv.Atoi(msg.Text)
	if err != nil {
		return fmt.Errorf("atoi[text=%v]: %w", msg.Text, err)
	}
	pe.smallestPeriod = time.Duration(days) * timeDay

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("delete message[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}
	pe.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pe.next(ctx, b, relatedMsgID, msg.Chat.ID, pe.SetBiggestPeriodMsg)
	if err != nil {
		return fmt.Errorf("next[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}

	return nil
}

func (pe *PeriodicEvent) SetBiggestPeriodMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "PeriodicEvent.SetBiggestPeriodMsg: %w"
	caption := pe.String() + "\n\nEnter biggest amount of days in period"
	pe.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pe.HandleMsgSetBiggestPeriod,
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

func (pe *PeriodicEvent) HandleMsgSetBiggestPeriod(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	days, err := strconv.Atoi(msg.Text)
	if err != nil {
		return fmt.Errorf("atoi[text=%v]: %w", msg.Text, err)
	}
	pe.biggestPeriod = time.Duration(days) * timeDay

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("delete message[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}
	pe.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pe.next(ctx, b, relatedMsgID, msg.Chat.ID, pe.SetTimeMsg)
	if err != nil {
		return fmt.Errorf("next[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}

	return nil
}

func (pe *PeriodicEvent) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetTimeMsg: %w"
	caption := pe.String() + "\n\nEnter time"

	pe.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pe.HandleMsgSetTime,
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

func (pe *PeriodicEvent) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetTime: %w"

	t, err := parseTime(msg.Text)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	pe.time = t
	pe.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pe.isWorkflow = false
	err = pe.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) SetDescription(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleEvent.SetDescription: %w"
	caption := pe.String() + "\n\nEnter description"

	pe.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pe.HandleMsgSetDescription,
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

func (pe *PeriodicEvent) HandleMsgSetDescription(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleEvent.HandleMsgSetDescription: %w"

	pe.description = msg.Text
	pe.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pe.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func computeStartTime(start time.Time, loc *time.Location) time.Duration {
	t := time.Date(0, 0, 0, start.Hour(), start.Minute(), 0, 0, loc)

	return t.Sub(t.Truncate(timeDay))
}

func (pe *PeriodicEvent) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	event := domains.PeriodicEvent{ //nolint:exhaustruct //no need to fill
		Text:               pe.text,
		Description:        pe.description,
		UserID:             user.ID,
		Start:              computeStartTime(pe.time, user.Location()),
		SmallestPeriod:     pe.smallestPeriod,
		BiggestPeriod:      pe.biggestPeriod,
		NotificationParams: nil,
	}

	_, err = pe.th.serv.AddPeriodicEvent(ctx, event, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pe.th.MainMenuWithText(ctx, b, msg, "Service successfully created:\n"+pe.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.UpdateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	params := service.UpdatePeriodicEventParams{
		ID:                 pe.id,
		Text:               pe.text,
		Description:        pe.description,
		UserID:             user.ID,
		Start:              computeStartTime(pe.time, user.Location()),
		SmallestPeriod:     pe.smallestPeriod,
		BiggestPeriod:      pe.biggestPeriod,
		NotificationParams: nil,
	}

	err = pe.th.serv.UpdatePeriodicEvent(ctx, params, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pe.th.MainMenuWithText(ctx, b, msg, "Service successfully updated:\n"+pe.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleEvent.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pe.th.serv.DeletePeriodicEvent(ctx, pe.id, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pe.th.MainMenuWithText(ctx, b, msg, "Service successfully deleted:\n"+pe.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pe *PeriodicEvent) HandleBtnEventChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsEventID []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	eventID, err := strconv.Atoi(string(btsEventID))
	if err != nil {
		return fmt.Errorf("strconv[string=%v]: %w", string(btsEventID), err)
	}

	event, err := pe.th.serv.GetPeriodicEvent(ctx, eventID, user.ID)
	if err != nil {
		return fmt.Errorf("get periodic event[eventID=%v,userID=%v]: %w", eventID, user.ID, err)
	}

	pe.id = event.ID
	pe.time = time.Time{}.Add(event.Start).In(user.Location())
	pe.text = event.Text
	pe.description = event.Description
	pe.isWorkflow = false
	pe.biggestPeriod = event.BiggestPeriod
	pe.smallestPeriod = event.SmallestPeriod

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, onSelectErrorHandling(pe.EditMenuMsg)).
		Button("Delete", nil, errorHandling(pe.DeleteInline)).
		Row().Button("Cancel", nil, errorHandling(pe.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     pe.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("edit message caption[chatID=%v,msgID=%v]: %w", msg.Chat.ID, msg.ID, err)
	}

	return nil
}
