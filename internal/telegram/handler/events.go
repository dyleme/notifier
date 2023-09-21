package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
		Row().Btn("Create event from task", createEvents.MessageChooseTask)

	return menu.Show
}

type ListEvents struct {
	serv *service.Service
}

func (l *ListEvents) list(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TelegramHandler.listEvents: %w"

	userInfo, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	events, err := l.serv.ListEvents(ctx, userInfo.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	if len(events) == 0 {
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
	for _, event := range events {
		// eventEdit := EventEdit{
		// 	serv: l.serv,
		// 	id:   event.ID,
		// 	text: event.Text,
		// 	time: time.Time{},
		// }
		kb.Row().Btn(event.Text, nil)
	}

	kbHandler, err := kb.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return kbHandler, nil
}

func NewEventCreation(serv *service.Service) EventCreation {
	return EventCreation{
		serv: serv,
		text: "",
		day:  time.Time{},
		time: time.Time{},
	}
}

type EventCreation struct {
	serv *service.Service
	text string
	day  time.Time
	time time.Time
}

func (ec *EventCreation) MessageChooseTask(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageChooseTask: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tasks, err := ec.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	menu := tgwf.NewMenuAction("Choose task")
	tgwf.AddSliceToMenu(menu, tasks, func(t domains.Task) string {
		return t.Text
	}, nil)

	_, err = menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ec.SetChosenTask, nil
}

func (ec *EventCreation) SetChosenTask(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "EventCreation.SetChosenTask: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	_, afterPoint, ok := strings.Cut(message.Text, ".")
	if !ok {
		return nil, fmt.Errorf(op, fmt.Errorf("no point"))
	}
	ec.text = afterPoint

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
	now := time.Now()
	tomorrow := time.Now().Add(day)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Set start day (in form 18.04)",
		ReplyMarkup: models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					models.KeyboardButton{Text: fmt.Sprintf("%02d.%02d", now.Day(), int(now.Month()))},
					models.KeyboardButton{Text: fmt.Sprintf("%02d.%02d", tomorrow.Day(), int(tomorrow.Month()))},
				},
			},
			ResizeKeyboard: true,
		},
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

func (ec *EventCreation) MessageSetStartTime(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.MessageSetStartTime: %w"
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

	return ec.create, nil
}

func computeTime(day, hourMinutes time.Time, loc *time.Location) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), hourMinutes.Hour(), hourMinutes.Minute(), 0, 0, loc)
}

func (ec *EventCreation) create(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventCreation.create: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
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

type EventEdit struct {
	serv *service.Service
	id   int
	text string
	day  time.Time
	time time.Time
}

func (ee *EventEdit) Menu(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventEdit.Menu: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	t, err := ee.serv.GetEvent(ctx, ee.id, user.ID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	message := fmt.Sprintf("Task:%s\nTime:%v", t.Text, t.Start.In(user.Location()).Format(time.DateTime))
	menu := tgwf.NewMenuAction(message).Row().
		Btn("Edit", ee.EditMenu).
		Btn("Delete", ee.Delete)

	menuHandler, err := menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return menuHandler, nil
}

func (ee *EventEdit) EditMenu(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "EventEdit.EditMenu: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	message := fmt.Sprintf("Task:%s\nTime:%v", ee.text, ee.time.In(user.Location()).Format(time.DateTime))
	menu := tgwf.NewMenuAction(message).Row().
		Btn("Edit text", ee.MessageSetText).
		Btn("Save", ee.save)

	menuHandler, err := menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return menuHandler, nil
}

func (ee *EventEdit) MessageSetText(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "SetTextAction.Show: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Enter task text",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ee.SetText, nil
}

func (ee *EventEdit) SetText(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "TaskEdit.SetText: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	ee.text = message.Text

	return ee.save, nil
}

func (ee *EventEdit) Delete(ctx context.Context, _ *bot.Bot, _ int64) (tgwf.Handler, error) {
	op := "EventEdit.Delete: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = ee.serv.DeleteEvent(ctx, ee.id, user.ID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}

func (ee *EventEdit) save(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TaskEdit.save: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = ee.serv.UpdateEvent(ctx, service.UpdateEventParams{
		ID:          ee.id,
		UserID:      user.ID,
		Description: "",
		Start:       computeTime(ee.time, ee.time, user.Location()),
		Done:        false,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Task successfully created",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}
