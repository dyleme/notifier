package handler

import (
	"context"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

const (
	timeDoublePointsFormat = "15:04"
	timeSpaceFormat        = "15 04"
)
const day = 24 * time.Hour

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

func makeOnSelect(f func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64)) func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	return func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
		f(ctx, b, msg.ID, msg.Chat.ID)
	}
}

func handleError(ctx context.Context, b *bot.Bot, chatID int64, err error) {
	log.Ctx(ctx).Error("error occurred", log.Err(err))
	if chatID == 0 {
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Server error occurred",
	})
	if err != nil {
		log.Ctx(ctx).Error("cannot send error message", log.Err(err))
	}
}
