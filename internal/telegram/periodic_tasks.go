package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/domain"
)

func (th *Handler) PeriodicTasksMenuInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "TelegramHandler.TasksMenuInline: %w"

	listTasks := ListPeriodicTasks{th: th}
	createTasks := NewPeriodicTaskCreation(th, true)
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("List periodic tasks", nil, errorHandling(listTasks.listInline)).
		Row().Button("Create periodic task", nil, onSelectErrorHandling(createTasks.SetTextMsg)).
		Row().Button("Cancel", nil, errorHandling(th.MainMenuInline))

	_, err := th.bot.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      mes.Chat.ID,
		MessageID:   mes.ID,
		Caption:     "Periodic tasks actions",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type ListPeriodicTasks struct {
	th *Handler
}

func (l *ListPeriodicTasks) listInline(ctx context.Context, b *bot.Bot, mes *models.Message, _ []byte) error {
	op := "ListTasks.listInline: %w"

	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	tasks, err := l.th.serv.ListPeriodicTasks(ctx, user.ID, defaultListParams)
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
		pt := PeriodicTask{th: l.th} //nolint:exhaustruct //fill it in pt.HandleBtnTaskChosen
		text := task.Text
		kbr.Row().Button(text, []byte(strconv.Itoa(task.ID)), errorHandling(pt.HandleBtnTaskChosen))
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

func NewPeriodicTaskCreation(th *Handler, isWorkflow bool) PeriodicTask {
	return PeriodicTask{
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

type PeriodicTask struct {
	th             *Handler
	id             int
	text           string
	smallestPeriod time.Duration
	biggestPeriod  time.Duration
	time           time.Time
	description    string
	isWorkflow     bool
}

func (pt *PeriodicTask) next(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64,
	nextFunc func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error,
) error {
	op := "SingleTask.next: %w"
	if pt.isWorkflow {
		err := nextFunc(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	} else {
		err := pt.EditMenuMsg(ctx, b, relatedMsgID, chatID)
		if err != nil {
			return fmt.Errorf(op, err)
		}
	}

	return nil
}

func (pt *PeriodicTask) isCreation() bool {
	return pt.id == notSettedID
}

func durToString(dur time.Duration) string {
	days := dur / timeDay
	if days == 0 {
		return ""
	}

	return strconv.Itoa(int(days))
}

func (pt *PeriodicTask) String() string {
	var (
		dateStr string
		timeStr string
	)

	if !pt.time.IsZero() {
		timeStr = pt.time.Format(timeDoublePointsFormat)
	}

	var taskStringBuilder strings.Builder
	taskStringBuilder.WriteString(fmt.Sprintf("Text: %q\n", pt.text))
	taskStringBuilder.WriteString(fmt.Sprintf("Date: %s\n", dateStr))
	taskStringBuilder.WriteString(fmt.Sprintf("Time: %s\n", timeStr))
	taskStringBuilder.WriteString(fmt.Sprintf("Smallest period: %v\n", durToString(pt.smallestPeriod)))
	taskStringBuilder.WriteString(fmt.Sprintf("Biggest period: %v\n", durToString(pt.biggestPeriod)))
	taskStringBuilder.WriteString(fmt.Sprintf("Description: %s\n", pt.description))

	return taskStringBuilder.String()
}

func (pt *PeriodicTask) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set text", nil, onSelectErrorHandling(pt.SetTextMsg)).
		Button("Set smallest period", nil, onSelectErrorHandling(pt.SetSmallestPeriodMsg)).
		Button("Set biggest period", nil, onSelectErrorHandling(pt.SetBiggestPeriodMsg)).
		Button("Set time", nil, onSelectErrorHandling(pt.SetTimeMsg)).
		Button("Set description", nil, onSelectErrorHandling(pt.SetDescription))

	kbr.Row()
	if pt.isCreation() {
		kbr.Button("Create", nil, errorHandling(pt.CreateInline))
	} else {
		kbr.Button("Update", nil, errorHandling(pt.UpdateInline))
	}

	kbr.Button("Cancel", nil, errorHandling(pt.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     pt.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf("edit message caption: %w", err)
	}

	return nil
}

func (pt *PeriodicTask) SetTextMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetTextMsg: %w"
	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to specify
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   "Enter task text",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pt.HandleMsgSetText,
		messageID: relatedMsgID,
	})

	return nil
}

func (pt *PeriodicTask) HandleMsgSetText(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetText: %w"
	pt.text = msg.Text

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pt.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pt.next(ctx, b, relatedMsgID, msg.Chat.ID, pt.SetTimeMsg)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pt *PeriodicTask) SetTimeMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetTimeMsg: %w"
	caption := pt.String() + "\n\nEnter time"

	pt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pt.HandleMsgSetTime,
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

func (pt *PeriodicTask) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetTime: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	t, err := parseTime(msg.Text, user.Location())
	if err != nil {
		return fmt.Errorf(op, err)
	}
	pt.time = t
	pt.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	pt.isWorkflow = false

	err = pt.next(ctx, b, relatedMsgID, msg.Chat.ID, pt.SetSmallestPeriodMsg)
	if err != nil {
		return fmt.Errorf("next[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}

	return nil
}

func (pt *PeriodicTask) SetSmallestPeriodMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "PeriodicTask.SetSmallestPeriodMsg: %w"
	caption := pt.String() + "\n\nEnter smallest amount of days in period"
	pt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pt.HandleMsgSetSmallestPeriod,
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

func (pt *PeriodicTask) HandleMsgSetSmallestPeriod(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	days, err := strconv.Atoi(msg.Text)
	if err != nil {
		return fmt.Errorf("atoi[text=%v]: %w", msg.Text, err)
	}
	pt.smallestPeriod = time.Duration(days) * timeDay

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("delete message[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}
	pt.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pt.next(ctx, b, relatedMsgID, msg.Chat.ID, pt.SetBiggestPeriodMsg)
	if err != nil {
		return fmt.Errorf("next[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}

	return nil
}

func (pt *PeriodicTask) SetBiggestPeriodMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	caption := pt.String() + "\n\nEnter biggest amount of days in period"
	pt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pt.HandleMsgSetBiggestPeriod,
		messageID: relatedMsgID,
	})

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    chatID,
		MessageID: relatedMsgID,
		Caption:   caption,
	})
	if err != nil {
		return fmt.Errorf("edit message caption: %w", err)
	}

	return nil
}

func (pt *PeriodicTask) HandleMsgSetBiggestPeriod(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	days, err := strconv.Atoi(msg.Text)
	if err != nil {
		return fmt.Errorf("atoi[text=%v]: %w", msg.Text, err)
	}
	pt.biggestPeriod = time.Duration(days) * timeDay

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf("delete message[msgID=%v,chatID=%v]: %w", msg.ID, msg.Chat.ID, err)
	}
	pt.th.waitingActionsStore.Delete(msg.Chat.ID)

	err = pt.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf("edit menu msg: %w", err)
	}

	return nil
}

func (pt *PeriodicTask) SetDescription(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "SingleTask.SetDescription: %w"
	caption := pt.String() + "\n\nEnter description"

	pt.th.waitingActionsStore.StoreDefDur(chatID, TextMessageHandler{
		handle:    pt.HandleMsgSetDescription,
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

func (pt *PeriodicTask) HandleMsgSetDescription(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "SingleTask.HandleMsgSetDescription: %w"

	pt.description = msg.Text
	pt.th.waitingActionsStore.Delete(msg.Chat.ID)

	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pt.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func computeStartTime(start time.Time, loc *time.Location) time.Duration {
	t := time.Date(0, 0, 0, start.Hour(), start.Minute(), 0, 0, loc)

	return t.Sub(t.Truncate(timeDay))
}

func (pt *PeriodicTask) CreateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	task := domain.NewPeriodicTask(
		domain.TaskCreationParams{
			Text:        pt.text,
			Description: pt.description,
			UserID:      user.ID,
			Start:       computeStartTime(pt.time, user.Location()),
		},
		pt.smallestPeriod,
		pt.biggestPeriod,
	)

	err = pt.th.serv.CreatePeriodicTask(ctx, task)
	if err != nil {
		return fmt.Errorf("create periodic task userID[%v]: %w", user.ID, err)
	}

	err = pt.th.MainMenuWithText(ctx, b, msg, "Service successfully created:\n"+pt.String())
	if err != nil {
		return fmt.Errorf("main menu: %w", err)
	}

	return nil
}

func (pt *PeriodicTask) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.UpdateInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	params := domain.NewPeriodicTask(
		domain.TaskCreationParams{
			ID:          pt.id,
			Text:        pt.text,
			Description: pt.description,
			UserID:      user.ID,
			Start:       computeStartTime(pt.time, user.Location()),
		},
		pt.smallestPeriod,
		pt.biggestPeriod,
	)

	err = pt.th.serv.UpdatePeriodicTask(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pt.th.MainMenuWithText(ctx, b, msg, "Service successfully updated:\n"+pt.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pt *PeriodicTask) DeleteInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "SingleTask.DeleteInline: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pt.th.serv.DeleteTask(ctx, pt.id, user.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = pt.th.MainMenuWithText(ctx, b, msg, "Service successfully deleted:\n"+pt.String())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (pt *PeriodicTask) HandleBtnTaskChosen(ctx context.Context, b *bot.Bot, msg *models.Message, btsTaskID []byte) error {
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("user from ctx: %w", err)
	}

	taskID, err := strconv.Atoi(string(btsTaskID))
	if err != nil {
		return fmt.Errorf("strconv[string=%v]: %w", string(btsTaskID), err)
	}

	task, err := pt.th.serv.GetPeriodicTask(ctx, taskID, user.ID)
	if err != nil {
		return fmt.Errorf("get periodic task[taskID=%v,userID=%v]: %w", taskID, user.ID, err)
	}

	pt.id = task.ID
	pt.time = time.Time{}.Add(task.Start).In(user.Location())
	pt.text = task.Text
	pt.description = task.Description
	pt.isWorkflow = false
	pt.biggestPeriod = task.BiggestPeriod()
	pt.smallestPeriod = task.SmallestPeriod()

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Button("Edit", nil, onSelectErrorHandling(pt.EditMenuMsg)).
		Button("Delete", nil, errorHandling(pt.DeleteInline)).
		Row().Button("Cancel", nil, errorHandling(pt.th.MainMenuInline))

	_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     pt.String(),
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf("edit message caption[chatID=%v,msgID=%v]: %w", msg.Chat.ID, msg.ID, err)
	}

	return nil
}
