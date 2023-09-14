package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"

	"github.com/Dyleme/Notifier/internal/lib/tgwf"
	domains "github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

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
	op := "EnableNotifications.Enable: %w"
	userID, err := UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = en.serv.SetDefaultNotificationParams(ctx, domains.NotificationParams{
		Period: defaultNotificationPeriod,
		Params: domains.Params{
			Telegram: int(chatID),
		},
		DalayedTill: nil,
	}, userID)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return nil, nil
}
