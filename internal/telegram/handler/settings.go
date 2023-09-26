package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Dyleme/Notifier/internal/lib/tgwf"
	domains "github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

func (th *TelegramHandler) SettingsMenu() tgwf.Action {
	timezoneSetting := &TimezoneSettings{userRepo: th.userRepo, zone: 0, isDST: false}
	menu := tgwf.NewMenuAction("Settings").
		Row().Btn("Notifications", th.NotificationMenu()).
		Row().Btn("Timezone", timezoneSetting.CurrentTime)

	return menu.Show
}

type TimezoneSettings struct {
	userRepo UserRepo
	zone     int
	isDST    bool
}

func (ts *TimezoneSettings) CurrentTime(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TimezoneSettings.CurrentTime: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	userLoc := time.FixedZone("Temp/Location", user.Zone*int(time.Hour/time.Second))
	utcTime := time.Now().In(time.UTC)
	userTime := utcTime.In(userLoc)
	h, m, _ := userTime.Clock()
	messageText := fmt.Sprintf("Your time: %02d:%02d,\nYour timezone: UTC%+02d\nDST:%v", h, m, user.Zone, user.IsDST)
	menu := tgwf.NewMenuAction(messageText).
		Row().Btn("Update", ts.SpecifyTimeMessage).
		Row().Btn("Ok", nil)

	menuHandler, err := menu.Show(ctx, b, chatID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return menuHandler, nil
}

func (ts *TimezoneSettings) SpecifyTimeMessage(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TimezoneSettings.Enable: %w"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Write your current time (18:04)",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ts.SetTime, nil
}

func (ts *TimezoneSettings) SetTime(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "TimezoneSettings.HandleMsgSetTime: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	userTime, err := parseTime(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	h, m, _ := userTime.Clock()
	ts.zone = getTimezone(time.Now().In(time.UTC), h, m)

	return ts.IsDSTMessage, nil
}

func (ts *TimezoneSettings) IsDSTMessage(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TimezoneSettings.IsDSTMessage: %w"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Is your country using DST?",
		ReplyMarkup: models.ReplyKeyboardMarkup{ //nolint:exhaustruct //no need to specify
			Keyboard: [][]models.KeyboardButton{
				{
					{ //nolint:exhaustruct //no need to specify
						Text: "true",
					},
					{ //nolint:exhaustruct //no need to specify
						Text: "false",
					},
				},
			},
			OneTimeKeyboard: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return ts.IsDst, nil
}

func (ts *TimezoneSettings) IsDst(_ context.Context, _ *bot.Bot, update *models.Update) (tgwf.Action, error) {
	op := "TimezoneSettings.HandleMsgSetTime: %w"
	message, err := tgwf.GetMessage(update)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	isDst, err := strconv.ParseBool(message.Text)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	ts.isDST = isDst

	return ts.done, nil
}

func (ts *TimezoneSettings) done(ctx context.Context, b *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TimezoneSettings.done: %w"

	err := ts.userRepo.UpdateUserTime(ctx, int(chatID), ts.zone, ts.isDST)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{ //nolint:exhaustruct //no need to specify
		ChatID: chatID,
		Text:   "Successfully updated",
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}

func getTimezone(utcTime time.Time, userHours, userMinutes int) int {
	utcTime = utcTime.Round(time.Hour)
	userToday := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(), userHours, userMinutes, 0, 0, time.UTC)
	userYesterday := userToday.Add(-day)
	userTomorrow := userToday.Add(day)

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

func (th *TelegramHandler) NotificationMenu() tgwf.Action {
	enableNotifications := EnableNotifications{serv: th.serv}
	menu := tgwf.NewMenuAction("Notifications").
		Row().Btn("Enable", enableNotifications.Enable)

	return menu.Show
}

type EnableNotifications struct {
	serv *service.Service
}

const defaultNotificationPeriod = 5 * time.Minute

func (en *EnableNotifications) Enable(ctx context.Context, _ *bot.Bot, chatID int64) (tgwf.Handler, error) {
	op := "TimezoneSettings.Enable: %w"
	user, err := UserFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = en.serv.SetDefaultNotificationParams(ctx, domains.NotificationParams{
		Period: defaultNotificationPeriod,
		Params: domains.Params{
			Telegram: int(chatID),
			Webhook:  "",
			Cmd:      "",
		},
		DalayedTill: nil,
	}, user.ID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}
