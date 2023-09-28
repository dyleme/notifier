package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func (th *TelegramHandler) Notify(ctx context.Context, notif domains.SendingNotification) error {
	op := "TelegramHandler.Notify: %w"
	kb := inKbr.New(th.bot).Button("Done", []byte(strconv.Itoa(notif.EventID)), errorHandling(th.markDone))
	_, err := th.bot.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID:      notif.Params.Params.Telegram,
		Text:        notif.Message,
		ReplyMarkup: kb,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (th *TelegramHandler) markDone(ctx context.Context, b *bot.Bot, mes *models.Message, data []byte) error {
	op := "TelegramHandler.markDone: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}
	eventID, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	event, err := th.serv.GetEvent(ctx, user.ID, eventID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = th.serv.UpdateEvent(ctx, service.EventUpdateParams{
		ID:          event.ID,
		UserID:      user.ID,
		Text:        event.Text,
		Description: event.Description,
		Start:       event.Start,
		Done:        true,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: mes.Chat.ID,
		Text:   "Event marked as done",
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
