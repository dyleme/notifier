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
	domains "github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

var ErrCantParseMessage = errors.New("cant parse message")

func (th *TelegramHandler) EventsMenu() tgwf.Action {
	listEvents := ListEvents{serv: th.serv}
	createEvents := EventCreation{serv: th.serv}
	menu := tgwf.NewMenuAction("Tasks action").
		Row().Btn("List tasks", listEvents.list).
		Row().Btn("Create task", createEvents.MessageSetText)

	return menu.Show
}

type ListEvents struct {
	serv *service.Service
}

func (l *ListEvents) list(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TelegramHandler.listTasks: %w"

	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tasks, err := l.serv.ListTimetableTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	if len(tasks) == 0 {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "No events",
		})

		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	kb := tgwf.NewMenuAction("Tasks")
	for _, t := range tasks {
		kb.Row().Btn(t.Text, nil)
	}

	return kb.Show(ctx, b, chatID)
}

type EventCreation struct {
	serv         *service.Service
	text         string
	requiredTime time.Duration
	day          time.Time
	time         time.Time
	periodic     bool
}

func (ec *EventCreation) MessageSetText(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "SetTextAction.Show: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Enter task text",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetText, nil
}

func (ec *EventCreation) SetText(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, err
	}
	ec.text = message.Text
	return ec.MessageSetStartDay, nil
}

func (ec *EventCreation) MessageSetStartDay(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetStartDay: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
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
		return nil, err
	}

	day, err := parseDay(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.day = day

	return ec.MessageSetStartTime, nil
}

const (
	dayPointFormat         = "02.01" // TODO: move to enum generator
	daySpaceFormat         = "02 01"
	dayPointWithYearFormat = "02.01.2006"
	daySpaceWithYearFormat = "02 01 2006"
)

var dayFormats = []string{dayPointFormat, daySpaceFormat, dayPointWithYearFormat, daySpaceWithYearFormat}

var firstYear = time.Time{}.Year()

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
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
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
		return nil, err
	}

	t, err := parseTime(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.time = t

	return ec.MessageSetRequiredTime, nil
}

const (
	timeDoublePointsFormat = "15:04" // TODO: move to enum generator
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
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
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
	return ec.MessageSetPeriodic, nil
}

func (ec *EventCreation) MessageSetPeriodic(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetPeriodic: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Is periodic",
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{Text: "true"},
					{Text: "false"},
				},
			},
			OneTimeKeyboard: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetPeriodic, nil
}

func (ec *EventCreation) SetPeriodic(ctx context.Context, b *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetPeriodic: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	periodic, err := strconv.ParseBool(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ec.periodic = periodic

	return ec.create, nil
}

func (ec *EventCreation) create(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.create: %w"
	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	task := domains.Task{
		UserID:       userID,
		Text:         ec.text,
		RequiredTime: ec.requiredTime,
		Periodic:     ec.periodic,
		Done:         false,
		Archived:     false,
	}
	event := domains.TimetableTask{
		UserID:      userID,
		Text:        ec.text,
		Description: "",
		Start:       time.Date(ec.day.Year(), ec.day.Month(), ec.day.Day(), ec.time.Hour(), ec.time.Minute(), 0, 0, time.Local),
	}

	_, err = ec.serv.CreateTimetableTask(ctx, task, event)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Task successfully created",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}
