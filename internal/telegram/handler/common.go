package handler

import (
	"context"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/log"
)

const (
	defaultListLimit = 100
)

var defaultListParams = service.ListParams{
	Offset: 0,
	Limit:  defaultListLimit,
}

const (
	timeDoublePointsFormat = "15:04"
	timeSpaceFormat        = "15 04"

	dayPointFormat         = "02.01"
	daySpaceFormat         = "02 01"
	dayPointWithYearFormat = "02.01.2006"
	daySpaceWithYearFormat = "02 01 2006"

	dayTimeFormat = "02.01.2006 15:04"
)
const timeDay = 24 * time.Hour

var timeFormats = []string{timeDoublePointsFormat, timeSpaceFormat}

func parseTime(dayString string, loc *time.Location) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.ParseInLocation(format, dayString, loc)
		if err == nil { // err eq nil
			return t, nil
		}
	}

	return time.Time{}, ErrCantParseMessage
}

var dayFormats = []string{dayPointFormat, daySpaceFormat, dayPointWithYearFormat, daySpaceWithYearFormat}

func parseDate(dayString string) (time.Time, error) {
	for _, format := range dayFormats {
		t, err := time.Parse(format, dayString)
		if err != nil {
			continue
		}

		if t.Year() != 0 {
			return t, nil
		}

		t = t.AddDate(time.Now().Year(), 0, 0)
		if t.Before(time.Now().Add(-2 * timeDay)) {
			t = t.AddDate(1, 0, 0)
		}

		return t, nil
	}

	return time.Time{}, ErrCantParseMessage
}

func onSelectErrorHandling(
	f func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error,
) func(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage, _ []byte) {
	return func(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage, _ []byte) {
		err := f(ctx, b, msg.Message.ID, msg.Message.Chat.ID)
		if err != nil {
			handleError(ctx, b, msg.Message.Chat.ID, err)
		}
	}
}

func errorHandling(f func(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error) func(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage, _ []byte) {
	return func(ctx context.Context, b *bot.Bot, msg models.MaybeInaccessibleMessage, bts []byte) {
		err := f(ctx, b, msg.Message, bts)
		if err != nil {
			handleError(ctx, b, msg.Message.Chat.ID, err)
		}
	}
}

func handleError(ctx context.Context, b *bot.Bot, chatID int64, err error) {
	log.Ctx(ctx).Error("error occurred", log.Err(err))
	if chatID == 0 {
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Server error occurred\n" + err.Error(),
	})
	if err != nil {
		log.Ctx(ctx).Error("cannot send error message", log.Err(err))
	}
}
