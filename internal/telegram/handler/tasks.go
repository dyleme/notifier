package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/tgwf"
	domains "github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

// func TaskMenuAction() tgwf.Action {
// 	listTasks := &ListTasks{}
// 	taskCreation := &TaskCreation{}
// 	tgwf.NewMenuAction("Tasks").Btn("list", listTasks.list)
// 	return menu.Show
// }

// func (serv *TelegramHandler) TaskMenu(ctx context.Context, _ *bot.Bot, message *models.Message, data []byte) {
// 	op := "TelegramHandler.taskMenu: %w"
// 	chatID := message.Chat.ID
// 	kb := inline.New(serv.bot)
// 	kb = kb.Row().Button("Tasks", nil, serv.TaskMenu)
// 	kb = kb.Row().Button("list", nil, serv.listTasks)
// 	kb = kb.Row().Button("Create", nil, serv.createTask)
//
// 	_, err := serv.bot.SendMessage(ctx, &bot.SendMessageParams{
// 		ChatID:      chatID,
// 		Text:        "EventChosen",
// 		ReplyMarkup: kb,
// 	})
// 	if err != nil {
// 		serv.handleError(ctx, chatID, fmt.Errorf(op, err))
// 		return
// 	}
// }

func (th *TelegramHandler) TaskMenu() tgwf.Action {
	listTasks := ListTasks{serv: th.serv}
	createTask := TaskCreation{serv: th.serv, text: ""}
	menu := tgwf.NewMenuAction("Tasks action").
		Row().Btn("List tasks", listTasks.list).
		Row().Btn("Create task", createTask.MessageSetText)

	return menu.Show
}

type ListTasks struct {
	serv *service.Service
}

func (l *ListTasks) list(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TelegramHandler.listTasks: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	tasks, err := l.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	kb := tgwf.NewMenuAction("Tasks")
	for i, t := range tasks {
		text := strconv.Itoa(i+1) + ". " + t.Text
		te := TaskEdit{serv: l.serv, id: t.ID, text: t.Text}
		kb.Row().Btn(text, te.Menu)
	}

	if len(tasks) == 0 {
		kb = kb.Row().Btn("No tasks. Create new task", nil)
	}

	kbHandler, err := kb.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return kbHandler, nil
}

func (l *ListTasks) Post(ctx context.Context, _ *bot.Bot, _ *models.Update) (tgwf.Action, error) {
	log.Ctx(ctx).Error("not implemented", "action", "ListTasks")

	return nil, nil
}

type TaskCreation struct {
	serv *service.Service
	text string
}

func (tc *TaskCreation) create(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TaskCreation.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = tc.serv.AddTask(ctx, domains.Task{ //nolint:exhaustruct //object creation request
		UserID:   user.ID,
		Text:     tc.text,
		Archived: false,
		Periodic: false,
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

func (tc *TaskCreation) MessageSetText(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "SetTextAction.Show: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Enter task text",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return tc.SetText, nil
}

func (tc *TaskCreation) SetText(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "TaskCreation.SetText: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	tc.text = message.Text

	return tc.create, nil
}

type TaskEdit struct {
	serv *service.Service
	id   int
	text string
}

func (te *TaskEdit) Menu(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TaskEdit.EventChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	t, err := te.serv.GetTask(ctx, te.id, user.ID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	message := fmt.Sprintf("Task:\n%s", t.Text)
	menu := tgwf.NewMenuAction(message).Row().
		Btn("Edit", te.MessageSetText).
		Btn("Delete", te.Delete)

	menuHandler, err := menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return menuHandler, nil
}

func (te *TaskEdit) MessageSetText(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "SetTextAction.Show: %w"
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Enter task text",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return te.SetText, nil
}

func (te *TaskEdit) Delete(ctx context.Context, _ *bot.Bot, _ int64) (tgwf.Handler, error) {
	op := "TaskEdit.Delete: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = te.serv.DeleteTask(ctx, te.id, user.ID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}

func (te *TaskEdit) SetText(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "TaskEdit.SetText: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}
	te.text = message.Text

	return te.save, nil
}

func (te *TaskEdit) save(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TaskEdit.save: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	err = te.serv.UpdateTask(ctx, domains.Task{
		ID:       te.id,
		UserID:   user.ID,
		Text:     te.text,
		Archived: false,
		Periodic: false,
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
