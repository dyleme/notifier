package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

//go:generate mockgen -destination=mocks/notification_job_mocks.go -package=mocks . Notifier
type Notifier interface {
	Add(ctx context.Context, n domains.SendingNotification) error
	Delete(ctx context.Context, id int) error
}

type NotifierJob struct {
	repo         Repository
	notifier     Notifier
	checkPeriod  time.Duration
	nextSendTime time.Time
	timer        *time.Timer
}

func NewNotifierJob(repo Repository, notifier Notifier, config Config) *NotifierJob {
	return &NotifierJob{
		repo:         repo,
		notifier:     notifier,
		checkPeriod:  config.CheckTasksPeriod,
		timer:        time.NewTimer(config.CheckTasksPeriod),
		nextSendTime: time.Now().Add(config.CheckTasksPeriod),
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

func smallestTime(ts ...time.Time) time.Time {
	var smallest time.Time
	for _, t := range ts {
		if smallest.IsZero() || t.Before(smallest) {
			smallest = t
		}
	}

	return smallest
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
				log.Ctx(ctx).Error("periodic event time", log.Err(fmt.Errorf(op, err)))
			}
		}
	}()

	wg.Wait()

	nearestTime = smallestTime(eventTime, periodicEventTime)
	if nearestTime.IsZero() {
		nearestTime = time.Now().Add(nj.checkPeriod)
	}

	nj.setNextNotificationTime(nearestTime)
}

func (nj *NotifierJob) setNextNotificationTime(t time.Time) {
	nj.nextSendTime = t
	nj.timer.Reset(time.Until(nj.nextSendTime))
}

func (s *Service) RunNotificationJob(ctx context.Context) {
	s.notifierJob.RunJob(ctx)
}

const parallelJobsAmount = 10

func (nj *NotifierJob) notify(ctx context.Context) {
	op := "Service.notify: %w"
	events, err := nj.repo.Events().ListEventsBefore(ctx, nj.nextSendTime)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("notify error", log.Err(fmt.Errorf(op, err)))
		}
	}

	wg, wgCtx := errgroup.WithContext(ctx)
	wg.SetLimit(parallelJobsAmount)
	for _, ev := range events {
		ev := ev
		wg.Go(func() error {
			return nj.notifyEvent(wgCtx, &ev)
		})
	}

	err = wg.Wait()
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(fmt.Errorf(op, err)))
	}

	nj.nextNotification(ctx)
}

func (nj *NotifierJob) notifyEvent(ctx context.Context, ev *domains.Event) error {
	op := "NotifierJob.notifyEvent: %w"
	if ev.NotificationParams == nil {
		defParams, err := nj.repo.DefaultNotificationParams().Get(ctx, ev.UserID)
		if err != nil {
			return fmt.Errorf(op, err)
		}

		ev.NotificationParams = &defParams
	}

	err := nj.notifier.Add(ctx, domains.SendingNotification{
		EventID:          ev.ID,
		UserID:           ev.UserID,
		Message:          ev.Text,
		Description:      ev.Description,
		Params:           *ev.NotificationParams,
		NotificationTime: ev.SendTime,
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = nj.repo.Events().MarkNotified(ctx, ev.ID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
