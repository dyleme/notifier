package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/tgwf"
	domains "github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

var ErrCantParseMessage = errors.New("cant parse message")

func (th *TelegramHandler) EventsMenu() tgwf.Action {
	listEvents := ListEvents{serv: th.serv}
	createEvents := NewEventCreation(th.serv)
	menu := tgwf.NewMenuAction("Events actions").
		Row().Btn("List events", listEvents.list).
		Row().Btn("Create event", createEvents.MessageSetText).
		Row().Btn("Create event from task", createEvents.MessageSetText)

	return menu.Show
}

type ListEvents struct {
	serv *service.Service
}

func (l *ListEvents) list(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TelegramHandler.listEvents: %w"

	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tasks, err := l.serv.ListEvents(ctx, userID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	if len(tasks) == 0 {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
			ChatID: chatID,
			Text:   "No events",
		})

		if err != nil {
			return nil, fmt.Errorf(op, err)
		}

		return nil, nil
	}

	kb := tgwf.NewMenuAction("Events")
	for _, t := range tasks {
		kb.Row().Btn(t.Text, nil)
	}

	kbHandler, err := kb.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return kbHandler, nil
}

func NewEventCreation(serv *service.Service) EventCreation {
	return EventCreation{
		serv:         serv,
		text:         "",
		requiredTime: 0,
		day:          time.Time{},
		time:         time.Time{},
	}
}

type EventCreation struct {
	serv         *service.Service
	text         string
	requiredTime time.Duration
	day          time.Time
	time         time.Time
}

func (ec *EventCreation) MessageChooseTask(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageChooseTask: %w"

	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tasks, err := ec.serv.ListUserTasks(ctx, userID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	menu := tgwf.NewMenuAction("Tasks")
	tgwf.AddSliceToMenu(menu, tasks, func(t domains.Task) string {
		return t.Text
	}, nil)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Choose task",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	menuHandler, err := menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return menuHandler, nil
}

func (ec *EventCreation) SetChosenTask(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetChosenTask: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.text = message.Text

	return ec.MessageSetStartDay, nil
}

func (ec *EventCreation) MessageSetText(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetText: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Enter event text",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetText, nil
}

func (ec *EventCreation) SetText(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetText: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.text = message.Text

	return ec.MessageSetStartDay, nil
}

func (ec *EventCreation) MessageSetStartDay(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetStartDay: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Set start day (in form 18.04)",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetStartDay, nil
}

func (ec *EventCreation) SetStartDay(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetStartDay: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	day, err := parseDay(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.day = day

	return ec.MessageSetStartTime, nil
}

const (
	dayPointFormat         = "02.01"
	daySpaceFormat         = "02 01"
	dayPointWithYearFormat = "02.01.2006"
	daySpaceWithYearFormat = "02 01 2006"
)

var dayFormats = []string{dayPointFormat, daySpaceFormat, dayPointWithYearFormat, daySpaceWithYearFormat}

func parseDay(dayString string) (time.Time, error) {
	for _, format := range dayFormats {
		t, err := time.Parse(format, dayString)
		if err != nil {
			continue
		}

		if t.Year() == 0 {
			t = t.AddDate(time.Now().Year(), 0, 0)
			if t.Before(time.Now()) {
				t = t.AddDate(1, 0, 0)
			}
		}

		return t, nil
	}

	return time.Time{}, ErrCantParseMessage
}

func (ec *EventCreation) MessageSetStartTime(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetStartDTime: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Set start time (in format 18 04)",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetStartTime, nil
}

func (ec *EventCreation) SetStartTime(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetStartTime: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	t, err := parseTime(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.time = t

	return ec.MessageSetRequiredTime, nil
}

const (
	timeDoublePointsFormat = "15:04"
	timeSpaceFormat        = "15 04"
)

var timeFormats = []string{timeDoublePointsFormat, timeSpaceFormat}

func parseTime(dayString string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, dayString)
		if err == nil { // err eq nil
			return t, nil
		}
	}

	return time.Time{}, ErrCantParseMessage
}

func (ec *EventCreation) MessageSetRequiredTime(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetRequiredTime: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Provide required time in minutes",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetRequiredTime, nil
}

func (ec *EventCreation) SetRequiredTime(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetRequiredTime: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	dur, err := strconv.Atoi(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	ec.requiredTime = time.Duration(dur) * time.Minute

	return ec.create, nil
}

func (ec *EventCreation) create(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.create: %w"
	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	event := domains.Event{ //nolint:exhaustruct // don't know id on creation
		UserID:      userID,
		Text:        ec.text,
		Description: "",
		Start:       time.Date(ec.day.Year(), ec.day.Month(), ec.day.Day(), ec.time.Hour(), ec.time.Minute(), 0, 0, time.UTC),
		Done:        false,
		Notification: domains.Notification{
			Sended:             false,
			NotificationParams: nil,
		},
	}

	_, err = ec.serv.CreateEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Event successfully created",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}
