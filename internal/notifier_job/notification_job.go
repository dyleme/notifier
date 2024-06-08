package notifier_job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

//go:generate mockgen -destination=mocks/notification_job_mocks.go -package=mocks . Notifier
type Notifier interface {
	Add(ctx context.Context, notif domains.SendingNotification) error
	Delete(ctx context.Context, notifID int) error
}

type Repository interface {
	DefaultNotificationParams() service.NotificationParamsRepository
	Notifications() service.NotificationsRepository
}

type NotifierJob struct {
	repo         Repository
	notifier     Notifier
	checkPeriod  time.Duration
	nextSendTime time.Time
	sendTimeMx   *sync.RWMutex
	timer        *time.Timer
	tr           *trManager.Manager
}

type Config struct {
	CheckTasksPeriod time.Duration
}

func New(repo Repository, notifier Notifier, config Config, tr *trManager.Manager) *NotifierJob {
	return &NotifierJob{
		repo:         repo,
		notifier:     notifier,
		checkPeriod:  config.CheckTasksPeriod,
		timer:        time.NewTimer(config.CheckTasksPeriod),
		nextSendTime: time.Now().Add(config.CheckTasksPeriod),
		sendTimeMx:   &sync.RWMutex{},
		tr:           tr,
	}
}

func (nj *NotifierJob) Run(ctx context.Context) {
	nj.setNextNotificationTime(ctx)
	for {
		select {
		case <-nj.timer.C:
			nj.notify(ctx)
			nj.setNextNotificationTime(ctx)
		case <-ctx.Done():
			nj.timer.Stop()

			return
		}
	}
}

func (nj *NotifierJob) UpdateWithTime(ctx context.Context, t time.Time) {
	nj.sendTimeMx.RLock()
	newTimeIsBeforeCurrent := t.Before(nj.nextSendTime)
	nj.sendTimeMx.RUnlock()

	if newTimeIsBeforeCurrent {
		nj.setNextNotificationTime(ctx)
	}
}

func (nj *NotifierJob) setNextNotificationTime(ctx context.Context) {
	nj.sendTimeMx.Lock()
	defer nj.sendTimeMx.Unlock()

	t := nj.nearestCheckTime(ctx)
	log.Ctx(ctx).Debug("next notification time", slog.Time("time", t))
	nj.nextSendTime = t
	nj.timer.Reset(time.Until(nj.nextSendTime))
}

func (nj *NotifierJob) nearestCheckTime(ctx context.Context) time.Time {
	var nearestTime time.Time
	notif, err := nj.repo.Notifications().GetNearest(ctx, time.Now())
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("get nearest notification", log.Err(err))
		}
		log.Ctx(ctx).Debug("no nearest notifications found")
		nearestTime = time.Now().Truncate(time.Minute).Add(nj.checkPeriod)
	} else {
		nearestTime = notif.SendTime
	}

	return nearestTime
}

func (nj *NotifierJob) notify(ctx context.Context) {
	err := nj.tr.Do(ctx, func(ctx context.Context) error {
		notifs, err := nj.repo.Notifications().ListNotSended(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("list not sended notifications: %w", err)
		}
		log.Ctx(ctx).Debug("found not sended notifications", slog.Any("notifications", utils.DtoSlice(notifs, func(n domains.Notification) int { return n.ID })))

		ids := make([]int, 0, len(notifs))
		for _, n := range notifs {
			ids = append(ids, n.ID)

			notificationParams, err := nj.getNotificationParams(ctx, n)
			if err != nil {
				return fmt.Errorf("get notification params: %w", err)
			}

			sendingNotification := domains.NewSendingNotification(n, notificationParams)
			err = nj.notifier.Add(ctx, sendingNotification)
			if err != nil {
				return fmt.Errorf("notifier add: %w", err)
			}
		}

		err = nj.repo.Notifications().MarkSended(ctx, ids)
		if err != nil {
			return fmt.Errorf("mark sended notifications: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", nj.nextSendTime))
	}
}

func (nj *NotifierJob) getNotificationParams(ctx context.Context, notif domains.Notification) (domains.NotificationParams, error) {
	if notif.Params != nil {
		return *notif.Params, nil
	}

	params, err := nj.repo.DefaultNotificationParams().Get(ctx, notif.UserID)
	if err != nil {
		return domains.NotificationParams{}, fmt.Errorf("get default notification params: %w", err)
	}

	return params, nil
}
