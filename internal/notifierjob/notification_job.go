package notifierjob

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm"
	"github.com/benbjohnson/clock"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

//go:generate mockgen -destination=mocks/notifier_mocks.go -package=mocks . Repository
type Notifier interface {
	Notify(ctx context.Context, notif domains.SendingEvent) error
}

//go:generate mockgen -destination=mocks/repository_mocks.go -package=mocks . Notifier
type Repository interface {
	Update(ctx context.Context, event domains.Event) error
	ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error)
	GetNearest(ctx context.Context) (domains.Event, error)
}

type NotifierJob struct {
	repo         Repository
	notifier     Notifier
	checkPeriod  time.Duration
	nextSendTime time.Time
	sendTimeMx   *sync.RWMutex
	timer        *clock.Timer
	tm           TxManager
	clock        clock.Clock
}

type Config struct {
	CheckTasksPeriod time.Duration
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) (err error)
}

func New(repo Repository, config Config, tr TxManager, nower clock.Clock) *NotifierJob {
	return &NotifierJob{
		repo:         repo,
		notifier:     nil,
		checkPeriod:  config.CheckTasksPeriod,
		nextSendTime: nower.Now().Add(config.CheckTasksPeriod),
		sendTimeMx:   &sync.RWMutex{},
		timer:        nower.Timer(config.CheckTasksPeriod),
		tm:           tr,
		clock:        nower,
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
	nj.timer.Reset(nj.clock.Until(t))
}

func (nj *NotifierJob) nearestCheckTime(ctx context.Context) time.Time {
	nextPeriodicInvocationTime := nj.clock.Now().Truncate(time.Minute).Add(nj.checkPeriod)
	event, err := nj.repo.GetNearest(ctx)
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("get nearest event", log.Err(err))
		}
		log.Ctx(ctx).Debug("no nearest events found")

		return nextPeriodicInvocationTime
	}

	return utils.MinTime(nextPeriodicInvocationTime, event.NextSend)
}

func (nj *NotifierJob) notify(ctx context.Context) {
	err := nj.tm.Do(ctx, func(ctx context.Context) error {
		events, err := nj.repo.ListNotSended(ctx, nj.nextSendTime)
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

			ev = ev.Rescheule(nj.clock.Now())

			err = nj.repo.Update(ctx, ev)
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
