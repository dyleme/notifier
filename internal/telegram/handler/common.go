package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/log"
)

const (
	defaultListLimit   = 100
	defaultTempMsgTime = 5 * time.Second
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
		if t.Before(time.Now().Add(-2 * day)) {
			t = t.AddDate(1, 0, 0)
		}

		return t, nil
	}

	return time.Time{}, ErrCantParseMessage
}

func onSelectErrorHandling(f func(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error) func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "onSelectErrorHandling: %w"

	return func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
		err := f(ctx, b, msg.ID, msg.Chat.ID)
		if err != nil {
			handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))
		}
	}
}

func errorHandling(f func(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) error) func(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) {
	op := "errorHandling: %w"

	return func(ctx context.Context, b *bot.Bot, msg *models.Message, bts []byte) {
		err := f(ctx, b, msg, bts)
		if err != nil {
			handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))
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
		Text:   "Server error occurred",
	})
	if err != nil {
		log.Ctx(ctx).Error("cannot send error message", log.Err(err))
	}
}

func SendTempMsg(ctx context.Context, b *bot.Bot, params *bot.SendMessageParams, dur time.Duration) error {
	op := "SendTempMsg: %w"
	msg, err := b.SendMessage(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	go func() {
		time.Sleep(dur)
		_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
		})
		if err != nil {
			handleError(ctx, b, msg.Chat.ID, fmt.Errorf(op, err))
		}
	}()

	return nil
}
