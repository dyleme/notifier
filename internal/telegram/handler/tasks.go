package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func (th *TelegramHandler) TasksMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.TasksMenuInline: %w"
	listTasks := ListTasks{th: th}
	createTask := SingleTask{th: th, id: notSettedID, text: ""}
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List tasks", nil, errorHandling(listTasks.listInline)).
		Row().Button("Create task", nil, onSelectErrorHandling(createTask.SetTextMsg))

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Tasks actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type ListTasks struct {
	th *TelegramHandler
}

func (l *ListTasks) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "ListTasks.listInline: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	tasks, err := l.th.serv.ListUserTasks(ctx, user.ID, service.ListParams{
		Offset: 0,
		Limit:  defaultListLimit,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for i, t := range tasks {
		text := strconv.Itoa(i+1) + ". " + t.Text
		te := SingleTask{th: l.th, id: t.ID, text: t.Text}
		kbr.Row().Button(text, nil, errorHandling(te.HandleBtnTaskChosen))
	}

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "All tasks",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type SingleTask struct {
	th   *TelegramHandler
	id   int
	text string
}

func (st *SingleTask) String() string {
	var eventStringBuilder strings.Builder
	eventStringBuilder.WriteString("Current event\n")
	eventStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", st.text))

	return eventStringBuilder.String()
}

func (st *SingleTask) isCreation() bool {
	return st.id == notSettedID
}

func (st *SingleTask) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set text", nil, onSelectErrorHandling(st.SetTextMsg))

	kbr.Row()
	if st.isCreation() {
		kbr.Button("Create", nil, errorHandling(st.CreateInline))
	} else {
		kbr.Button("Update", nil, errorHandling(st.UpdateInline))
	}

	kbr.Button("Cancel", nil, errorHandling(st.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     st.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (st *SingleTask) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = st.th.serv.AddTask(ctx, domains.Task{ //nolint:exhaustruct //object creation request
		UserID:   user.ID,
		Text:     st.text,
		Archived: false,
		Periodic: false,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = st.th.MainMenuWithText(ctx, b, msg, "Task successfully created\n"+st.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (st *SingleTask) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter task text",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	st.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    st.HandleMsgSetText,
		messageID: relatedMsgID,
	})

	return nil
}

func (st *SingleTask) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetText: %w"

	st.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	st.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = st.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (st *SingleTask) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TaskEdit.EventChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	t, err := st.th.serv.GetTask(ctx, st.id, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	st.id = t.ID
	st.text = t.Text

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Edit", nil, onSelectErrorHandling(st.EditMenuMsg)).
		Row().Button("Delete", nil, errorHandling(st.DeleteInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     st.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (st *SingleTask) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = st.th.serv.DeleteTask(ctx, st.id, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = st.th.MainMenuWithText(ctx, b, msg, "Task successfully deleted\n"+st.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (st *SingleTask) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TaskEdit.save: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = st.th.serv.UpdateTask(ctx, domains.Task{
		ID:       st.id,
		UserID:   user.ID,
		Text:     st.text,
		Archived: false,
		Periodic: false,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = st.th.MainMenuWithText(ctx, b, msg, "Task successfully updated\n"+st.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
