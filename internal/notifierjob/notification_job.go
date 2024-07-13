package notifierjob

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

//go:generate mockgen -destination=mocks/event_job_mocks.go -package=mocks . Notifier
type Notifier interface {
	Notify(ctx context.Context, notif domains.SendingEvent) error
}

type Repository interface {
	DefaultEventParams() service.NotificationParamsRepository
	Events() service.EventsRepository
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

func New(repo Repository, config Config, tr *trManager.Manager) *NotifierJob {
	return &NotifierJob{
		repo:         repo,
		notifier:     nil,
		checkPeriod:  config.CheckTasksPeriod,
		nextSendTime: time.Now().Add(config.CheckTasksPeriod),
		sendTimeMx:   &sync.RWMutex{},
		timer:        time.NewTimer(config.CheckTasksPeriod),
		tr:           tr,
	}
}

func (nj *NotifierJob) SetNotifier(notifier Notifier) {
	nj.notifier = notifier
}

func (nj *NotifierJob) Run(ctx context.Context) {
	nj.setNextEventTime(ctx)
	for {
		select {
		case <-nj.timer.C:
			nj.notify(ctx)
			nj.setNextEventTime(ctx)
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
		nj.setNextEventTime(ctx)
	}
}

func (nj *NotifierJob) setNextEventTime(ctx context.Context) {
	nj.sendTimeMx.Lock()
	defer nj.sendTimeMx.Unlock()

	t := nj.nearestCheckTime(ctx)
	log.Ctx(ctx).Debug("next event time", slog.Time("time", t))
	nj.nextSendTime = t
	nj.timer.Reset(time.Until(nj.nextSendTime))
}

func (nj *NotifierJob) nearestCheckTime(ctx context.Context) time.Time {
	nextPeriodicInvocationTime := time.Now().Truncate(time.Minute).Add(nj.checkPeriod)
	event, err := nj.repo.Events().GetNearest(ctx)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("get nearest event", log.Err(err))
		}
		log.Ctx(ctx).Debug("no nearest events found")

		return nextPeriodicInvocationTime
	}

	return utils.MinTime(nextPeriodicInvocationTime, event.NextSendTime)
}

func (nj *NotifierJob) notify(ctx context.Context) {
	err := nj.tr.Do(ctx, func(ctx context.Context) error {
		events, err := nj.repo.Events().ListNotSended(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("list not sended events: %w", err)
		}
		log.Ctx(ctx).Info("found not sended events", slog.Any("events", utils.DtoSlice(events, func(n domains.Event) int { return n.ID })))

		for _, ev := range events {
			sendingEvent := domains.NewSendingEvent(ev)
			err := nj.notifier.Notify(ctx, sendingEvent)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))
			}

			ev = ev.Rescheule()

			err = nj.repo.Events().Update(ctx, ev)
			if err != nil {
				log.Ctx(ctx).Error("update event", log.Err(err), slog.Any("event", ev))
			}
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", nj.nextSendTime))
	}
}
