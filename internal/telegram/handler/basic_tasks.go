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

func (th *TelegramHandler) TasksMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.TasksMenuInline: %w"

	listTasks := ListTasks{th: th}
	createTasks := NewTaskCreation(th, true)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List tasks", nil, errorHandling(listTasks.listInline)).
		Row().Button("Create task", nil, onSelectErrorHandling(createTasks.SetTextMsg)).
		Row().Button("Cancel", nil, errorHandling(th.MainMenuInline))

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

	tasks, err := l.th.serv.ListBasicTasks(ctx, user.ID, service.ListFilterParams{
		ListParams: defaultListParams,
		TagIDs:     []int{},
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if len(tasks) == 0 {
		kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
		kbr.Row().Button("Ok", nil, errorHandling(l.th.MainMenuInline))
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
			ChatID:    mes.Chat.ID,
			MessageID: mes.ID,
			Caption:   "No tasks",
		})
		if err != nil {
			return fmt.Errorf(op, err)
		}

		return nil
	}

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick())
	for _, task := range tasks {
		ec := BasicTask{th: l.th} //nolint:exhaustruct //fill it in ec.HandleBtnTaskChosen
		text := task.Text + "\t|\t" + task.Start.In(user.Location()).Format(dayTimeFormat)
		kbr.Row().Button(text, []byte(strconv.Itoa(task.ID)), errorHandling(ec.HandleBtnTaskChosen))
	}
	kbr.Row().Button("Cancel", nil, errorHandling(l.th.MainMenuInline))

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

const notSettedID = -1

func NewTaskCreation(th *TelegramHandler, isWorkflow bool) BasicTask {
	return BasicTask{
		id:          notSettedID,
		th:          th,
		text:        "",
		description: "",
		date:        time.Time{},
		time:        time.Time{},
		isWorkflow:  isWorkflow,
	}
}

type BasicTask struct {
	th          *TelegramHandler
	id          int
	text        string
	date        time.Time
	time        time.Time
	description string
	isWorkflow  bool
}

func (bt *BasicTask) next(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64,
	nextFunc func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error,
) error {
	op := "SingleTask.next: %w"
	if bt.isWorkflow {
		err := nextFunc(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	} else {
		err := bt.EditMenuMsg(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	}

	return nil
}

func (bt *BasicTask) isCreation() bool {
	return bt.id == notSettedID
}

func (bt *BasicTask) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !bt.date.IsZero() {
		dateStr = bt.date.Format(dayPointWithYearFormat)
	}
	if !bt.time.IsZero() {
		timeStr = bt.time.Format(timeDoublePointsFormat)
	}

	var taskStringBuilder strings.Builder
	taskStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", bt.text))
	taskStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	taskStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))
	taskStringBuilder.WriteString(fmt.Sprintf("Description: %s\n", bt.description))

	return taskStringBuilder.String()
}

func (bt *BasicTask) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "BasicTask.EditMenuMsg: %w"
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set text", nil, onSelectErrorHandling(bt.SetTextMsg)).
		Button("Set date", nil, onSelectErrorHandling(bt.SetDateMsg)).
		Button("Set time", nil, onSelectErrorHandling(bt.SetTimeMsg)).
		Button("Set description", nil, onSelectErrorHandling(bt.SetDescription))

	kbr.Row()
	if bt.isCreation() {
		kbr.Button("Create", nil, errorHandling(bt.CreateInline))
	} else {
		kbr.Button("Update", nil, errorHandling(bt.UpdateInline))
	}

	kbr.Button("Cancel", nil, errorHandling(bt.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     bt.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter task text",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	bt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    bt.HandleMsgSetText,
		messageID: relatedMsgID,
	})

	return nil
}

func (bt *BasicTask) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetText: %w"
	bt.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	bt.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = bt.next(ctx, b, relatedMsgID, msg.Chat.ID, bt.SetDateMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) SetDateMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "BasicTask.SetDateMsg: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	caption := bt.String() + "\n\nEnter date (it can bt or one of provided, or you can type your own date)"
	now := time.Now().In(user.Location())
	nowStr := now.Format(dayPointFormat)
	tomorrow := time.Now().Add(timeDay).In(user.Location())
	tomorrowStr := tomorrow.Format(dayPointFormat)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button(nowStr, []byte(nowStr), errorHandling(bt.HandleBtnSetDate)).
		Row().Button(tomorrowStr, []byte(tomorrowStr), errorHandling(bt.HandleBtnSetDate))

	bt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    bt.HandleMsgSetDate,
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

func (bt *BasicTask) HandleBtnSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error {
	op := "BasicTask.HandleBtnSetDate: %w"

	if err := bt.handleSetDate(ctx, b, msg.Chat.ID, msg.ID, string(bts)); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) HandleMsgSetDate(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetDate: %w"

	if err := bt.handleSetDate(ctx, b, msg.Chat.ID, relatedMsgID, msg.Text); err != nil {
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

func (bt *BasicTask) handleSetDate(ctx context.Context, b *bot.Bot, chatID int64, msgID int, dateStr string) error {
	op := "SingleTask.handleSetDate: %w"

	t, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	bt.date = t

	bt.th.waitingActionsStore.Delete(chatID)

	err = bt.next(ctx, b, msgID, chatID, bt.SetTimeMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetTimeMsg: %w"
	caption := bt.String() + "\n\nEnter time"

	bt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    bt.HandleMsgSetTime,
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

func (bt *BasicTask) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetTime: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	t, err := parseTime(msg.Text, user.Location())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	bt.time = t
	bt.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	bt.isWorkflow = false
	err = bt.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) SetDescription(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetDescription: %w"
	caption := bt.String() + "\n\nEnter description"

	bt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    bt.HandleMsgSetDescription,
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

func (bt *BasicTask) HandleMsgSetDescription(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetDescription: %w"

	bt.description = msg.Text
	bt.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = bt.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

var ErrTimeInPast = errors.New("time is in past")

func (bt *BasicTask) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.CreateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	t := bt.date.Add(bt.time.Sub(bt.time.Truncate(timeDay)))

	if t.Before(time.Now()) {
		return fmt.Errorf(op, ErrTimeInPast)
	}

	task := domains.BasicTask{ //nolint:exhaustruct // don't know id on creation
		UserID:             user.ID,
		Text:               bt.text,
		Description:        bt.description,
		Start:              t,
		NotificationParams: nil,
	}

	_, err = bt.th.serv.CreateBasicTask(ctx, task)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = bt.th.MainMenuWithText(ctx, b, msg, "Service successfully created:\n"+bt.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.UpdateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	t := bt.date.Add(bt.time.Sub(bt.time.Truncate(timeDay)))

	_, err = bt.th.serv.UpdateBasicTask(ctx, domains.BasicTask{
		ID:                 bt.id,
		Text:               bt.text,
		UserID:             user.ID,
		Description:        bt.description,
		Start:              t,
		NotificationParams: nil,
		Tags:               nil,
	}, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = bt.th.MainMenuWithText(ctx, b, msg, "Service successfully updated:\n"+bt.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = bt.th.serv.DeleteBasicTask(ctx, user.ID, bt.id)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = bt.th.MainMenuWithText(ctx, b, msg, "Service successfully deleted:\n"+bt.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (bt *BasicTask) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsTaskID []byte) error {
	op := "TaskEdit.TaskChosen: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	taskID, err := strconv.Atoi(string(btsTaskID))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	task, err := bt.th.serv.GetBasicTask(ctx, user.ID, taskID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	bt.id = task.ID
	bt.date = task.Start
	bt.time = task.Start
	bt.text = task.Text
	bt.description = task.Description
	bt.isWorkflow = false

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, onSelectErrorHandling(bt.EditMenuMsg)).
		Button("Delete", nil, errorHandling(bt.DeleteInline)).
		Row().Button("Cancel", nil, errorHandling(bt.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     bt.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
