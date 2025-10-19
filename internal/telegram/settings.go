package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	inKbr "github.com/go-telegram/ui/keyboard/inline"

	"github.com/dyleme/Notifier/internal/domain"
)

func (th *Handler) SettingsInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TelegramHandler.SettingsInline: %w"

	timezoneSetting := &TimezoneSettings{th: th, zone: 0, isDST: false}
	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().Button("Timezone", nil, errorHandling(timezoneSetting.CurrentTime))

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     "Timezone",
		ReplyMarkup: kbr,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

type TimezoneSettings struct {
	th    *Handler
	zone  int
	isDST bool
}

func (ts *TimezoneSettings) String() string {
	userLoc := time.FixedZone("Temp/Location", ts.zone*int(time.Hour/time.Second))
	utcTime := time.Now().In(time.UTC)
	userTime := utcTime.In(userLoc)
	h, m, _ := userTime.Clock()

	return fmt.Sprintf("Your time: %02d:%02d,\nYour timezone: UTC%+02d\nDST:%v", h, m, ts.zone, ts.isDST)
}

func (ts *TimezoneSettings) CurrentTime(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TimezoneSettings.CurrentTime: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	ts.isDST = user.IsTimeZoneDST
	ts.zone = user.TimeZoneOffset

	if err = ts.EditMenuMsg(ctx, b, msg.ID, msg.Chat.ID); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ts *TimezoneSettings) EditMenuMsg(ctx context.Context, b *bot.Bot, relatedMsgID int, chatID int64) error {
	op := "TimezoneSettings.EditMenuMsg: %w"

	kbr := inKbr.New(b, inKbr.NoDeleteAfterClick()).
		Row().
		Button("Set current time", nil, errorHandling(ts.SetTimeMsg)).
		Button("Set is dst", nil, errorHandling(ts.SetDstMsg))

	kbr.Row().Button("Update", nil, errorHandling(ts.UpdateInline))

	kbr.Button("Cancel", nil, errorHandling(ts.th.MainMenuInline))

	params := &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      chatID,
		MessageID:   relatedMsgID,
		Caption:     ts.String(),
		ReplyMarkup: kbr,
	}

	_, err := b.EditMessageCaption(ctx, params)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ts *TimezoneSettings) SetTimeMsg(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TimezoneSettings.SetTimeMsg: %w"

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Caption:     ts.String() + "\n\n" + "Specify time message",
		ReplyMarkup: inKbr.New(b, inKbr.NoDeleteAfterClick()).Button("Cancel", nil, errorHandling(ts.th.MainMenuInline)),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	ts.th.waitingActionsStore.StoreDefDur(msg.Chat.ID, TextMessageHandler{
		handle:    ts.HandleMsgSetTime,
		messageID: msg.ID,
	})

	return nil
}

func (ts *TimezoneSettings) HandleMsgSetTime(ctx context.Context, b *bot.Bot, msg *models.Message, relatedMsgID int) error {
	op := "TimezoneSettings.HandleMsgSetTime: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	userTime, err := parseTime(msg.Text, user.Location())
	if err != nil {
		return fmt.Errorf(op, err)
	}

	h, m, _ := userTime.Clock()
	ts.zone = getTimezone(time.Now().In(time.UTC), h, m)

	err = ts.EditMenuMsg(ctx, b, relatedMsgID, msg.Chat.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ts *TimezoneSettings) SetDstMsg(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TimezoneSettings.SetDstMsg: %w"

	_, err := b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{ //nolint:exhaustruct //no need to fill
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		Caption:   "Is your country using DST?",
		ReplyMarkup: inKbr.New(b, inKbr.NoDeleteAfterClick()).Row().
			Button("true", []byte("true"), errorHandling(ts.HandleBtnSetDst)).
			Button("false", []byte("false"), errorHandling(ts.HandleBtnSetDst)),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ts *TimezoneSettings) HandleBtnSetDst(ctx context.Context, b *bot.Bot, msg *models.Message, boolBts []byte) error {
	op := "TimezoneSettings.HandleBtnSetDst: %w"

	isDst, err := strconv.ParseBool(string(boolBts))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	ts.isDST = isDst

	if err = ts.EditMenuMsg(ctx, b, msg.ID, msg.Chat.ID); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (ts *TimezoneSettings) UpdateInline(ctx context.Context, b *bot.Bot, msg *models.Message, _ []byte) error {
	op := "TimezoneSettings.UpdateInline: %w"

	err := ts.th.serv.UpdateUserTime(ctx, int(msg.Chat.ID), domain.TimeZoneOffset(ts.zone), ts.isDST)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	if err = ts.th.MainMenuWithText(ctx, b, msg, "Timezone updated:"+ts.String()); err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func getTimezone(utcTime time.Time, userHours, userMinutes int) int {
	utcTime = utcTime.Round(time.Hour)
	userToday := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(), userHours, userMinutes, 0, 0, time.UTC)
	userYesterday := userToday.Add(-timeDay)
	userTomorrow := userToday.Add(timeDay)

	realUserTime := userYesterday
	minDiff := absDur(utcTime.Sub(userYesterday))

	if diff := absDur(utcTime.Sub(userToday)); diff < minDiff {
		realUserTime = userToday
		minDiff = diff
	}

	if diff := absDur(utcTime.Sub(userTomorrow)); diff < minDiff {
		realUserTime = userTomorrow
	}

	realUserTime = realUserTime.Round(time.Hour)

	difference := realUserTime.Sub(utcTime) / time.Hour

	return int(difference)
}

func absDur(d time.Duration) time.Duration {
	if d > 0 {
		return d
	}

	return -d
}
