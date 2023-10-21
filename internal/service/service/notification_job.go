package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	trManager "github.com/avito-tech/go-transaction-manager/trm/manager"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

//go:generate mockgen -destination=mocks/notification_job_mocks.go -package=mocks . Notifier
type Notifier interface {
	Add(ctx context.Context, notif domains.SendingNotification) error
	Delete(ctx context.Context, eventID, userID int) error
}

type NotifierJob struct {
	repo         Repository
	notifier     Notifier
	checkPeriod  time.Duration
	nextSendTime time.Time
	timer        *time.Timer
	tr           *trManager.Manager
}

func NewNotifierJob(repo Repository, notifier Notifier, config Config, tr *trManager.Manager) *NotifierJob {
	return &NotifierJob{
		repo:         repo,
		notifier:     notifier,
		checkPeriod:  config.CheckTasksPeriod,
		timer:        time.NewTimer(config.CheckTasksPeriod),
		nextSendTime: time.Now().Add(config.CheckTasksPeriod),
		tr:           tr,
	}
}

func (nj *NotifierJob) RunJob(ctx context.Context) {
	nj.notify(ctx)
	for {
		select {
		case <-nj.timer.C:
			nj.notify(ctx)
		case <-ctx.Done():
			nj.timer.Stop()

			return
		}
	}
}

func (nj *NotifierJob) UpdateWithTime(ctx context.Context, t time.Time) {
	if t.Before(nj.nextSendTime) {
		nj.nextNotification(ctx)
	}
}

func (nj *NotifierJob) nextNotification(ctx context.Context) {
	op := "NotifierJob.nextNotification: %w"
	var nearestTime time.Time
	var (
		wg                sync.WaitGroup
		eventTime         time.Time
		periodicEventTime time.Time
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		eventTime, err = nj.repo.Events().GetNearestEventSendTime(ctx)
		if err != nil {
			var notFoundErr serverrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				log.Ctx(ctx).Error("event time", log.Err(fmt.Errorf(op, err)))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		periodicEventTime, err = nj.repo.PeriodicEvents().GetNearestNotificationSendTime(ctx)
		if err != nil {
			var notFoundErr serverrors.NotFoundError
			if !errors.As(err, &notFoundErr) {
				log.Ctx(ctx).Error("periodic event time", log.Err(fmt.Errorf("get nearest notification send time: %w", err)))
			}
		}
	}()

	wg.Wait()
	log.Ctx(ctx).Debug("nearest event times", "basicEvent", eventTime, "periodcEvent", periodicEventTime)

	nearestTime = slices.MinFunc([]time.Time{eventTime, periodicEventTime}, utils.TimeCmpWithoutZero)
	if nearestTime.Before(time.Now()) {
		nearestTime = time.Now().Add(nj.checkPeriod)
	}

	nj.setNextNotificationTime(ctx, nearestTime)
}

func (nj *NotifierJob) setNextNotificationTime(ctx context.Context, t time.Time) {
	log.Ctx(ctx).Debug("notifier service new time", "time", t)
	nj.nextSendTime = t
	nj.timer.Reset(time.Until(nj.nextSendTime))
}

func (s *Service) RunNotificationJob(ctx context.Context) {
	s.notifierJob.RunJob(ctx)
}

func (nj *NotifierJob) notifiedBasicEvents(ctx context.Context) []domains.SendingNotification {
	events, err := nj.repo.Events().ListEventsBefore(ctx, nj.nextSendTime)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("notified basic events", log.Err(fmt.Errorf("list events before: %w", err)))
		}
	}

	notifs := make([]domains.SendingNotification, 0, len(events))
	for _, ev := range events {
		if ev.NotificationParams == nil {
			params, err := nj.repo.DefaultNotificationParams().Get(ctx, ev.UserID)
			if err != nil {
				log.Ctx(ctx).Error("notified basic events", log.Err(fmt.Errorf("default notification params get: %w", err)))

				continue
			}

			ev.NotificationParams = &params
		}
		notifs = append(notifs, domains.SendingNotification{
			EventType:        domains.BasicEventType,
			EventID:          ev.ID,
			UserID:           ev.UserID,
			Message:          ev.Text,
			Description:      ev.Description,
			Params:           *ev.NotificationParams,
			NotificationTime: ev.SendTime,
		})
	}

	return notifs
}

func (nj *NotifierJob) notifiedPeriodicEvents(ctx context.Context) []domains.SendingNotification {
	events, err := nj.repo.PeriodicEvents().ListNotificationsAtSendTime(ctx, nj.nextSendTime)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("notified basic events", log.Err(fmt.Errorf("list events before: %w", err)))
		}
	}

	notifs := make([]domains.SendingNotification, 0, len(events))
	for _, ev := range events {
		if ev.NotificationParams == nil {
			params, err := nj.repo.DefaultNotificationParams().Get(ctx, ev.UserID)
			if err != nil {
				log.Ctx(ctx).Error("notified basic events", log.Err(fmt.Errorf("default notification params get: %w", err)))

				continue
			}

			ev.NotificationParams = &params
		}
		notifs = append(notifs, domains.SendingNotification{
			EventType:        domains.PeriodicEventType,
			EventID:          ev.ID,
			UserID:           ev.UserID,
			Message:          ev.Text,
			Description:      ev.Description,
			Params:           *ev.NotificationParams,
			NotificationTime: ev.Notification.SendTime,
		})
	}

	return notifs
}

func (nj *NotifierJob) getEventsToNotify(ctx context.Context) []domains.SendingNotification {
	var (
		wg             sync.WaitGroup
		basicNotifs    []domains.SendingNotification
		periodicNotifs []domains.SendingNotification
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		basicNotifs = nj.notifiedBasicEvents(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		periodicNotifs = nj.notifiedPeriodicEvents(ctx)
	}()
	wg.Wait()

	return append(basicNotifs, periodicNotifs...)
}

func (nj *NotifierJob) notify(ctx context.Context) {
	ns := nj.getEventsToNotify(ctx)
	log.Ctx(ctx).Debug("notifcation service notification", "", fmt.Sprintf("%#v", ns))
	for i := 0; i < len(ns); i++ {
		err := nj.send(ctx, ns[i])
		if err != nil {
			log.Ctx(ctx).Error("send", log.Err(fmt.Errorf("send[eventID=%v,eventType=%v]: %w", ns[i].EventID, ns[i].EventType, err)))
		}
	}

	nj.nextNotification(ctx)
}

func (nj *NotifierJob) send(ctx context.Context, notif domains.SendingNotification) error {
	err := nj.markNotified(ctx, notif)
	if err != nil {
		return fmt.Errorf("mark notified: %w", err)
	}

	err = nj.notifier.Add(ctx, notif)
	if err != nil {
		return fmt.Errorf("notifer add: %w", err)
	}

	return nil
}

func (nj *NotifierJob) markNotified(ctx context.Context, notif domains.SendingNotification) error {
	switch notif.EventType {
	case domains.BasicEventType:
		err := nj.repo.Events().MarkNotified(ctx, notif.EventID)
		if err != nil {
			return fmt.Errorf("mark notified: %w", err)
		}
	case domains.PeriodicEventType:
		err := nj.tr.Do(ctx, func(ctx context.Context) error {
			ev, err := nj.repo.PeriodicEvents().Get(ctx, notif.EventID, notif.UserID)
			if err != nil {
				return fmt.Errorf("get[eventID=%v,userID=%v]: %w", notif.EventID, notif.UserID, err)
			}

			err = nj.repo.PeriodicEvents().MarkNotificationSend(ctx, ev.Notification.ID)
			if err != nil {
				return fmt.Errorf("mark notified: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("atomic: %w", err)
		}
	default:
		return fmt.Errorf("invalid event type")
	}

	return nil
}
