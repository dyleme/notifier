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
	Add(ctx context.Context, event domains.SendingEvent) error
	Delete(ctx context.Context, eventID int) error
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
	var nearestTime time.Time
	event, err := nj.repo.Events().GetNearest(ctx, time.Now())
	if err != nil {
		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("get nearest event", log.Err(err))
		}
		log.Ctx(ctx).Debug("no nearest events found")
		nearestTime = time.Now().Truncate(time.Minute).Add(nj.checkPeriod)
	} else {
		nearestTime = event.SendTime
	}

	return nearestTime
}

func (nj *NotifierJob) notify(ctx context.Context) {
	err := nj.tr.Do(ctx, func(ctx context.Context) error {
		notifs, err := nj.repo.Events().ListNotSended(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("list not sended events: %w", err)
		}
		log.Ctx(ctx).Debug("found not sended events", slog.Any("events", utils.DtoSlice(notifs, func(n domains.Event) int { return n.ID })))

		ids := make([]int, 0, len(notifs))
		for _, n := range notifs {
			ids = append(ids, n.ID)

			eventParams, err := nj.getEventParams(ctx, n)
			if err != nil {
				return fmt.Errorf("get event params: %w", err)
			}

			sendingEvent := domains.NewSendingEvent(n, eventParams)
			err = nj.notifier.Add(ctx, sendingEvent)
			if err != nil {
				return fmt.Errorf("notifier add: %w", err)
			}
		}

		err = nj.repo.Events().MarkSended(ctx, ids)
		if err != nil {
			return fmt.Errorf("mark sended events: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", nj.nextSendTime))
	}
}

func (nj *NotifierJob) getEventParams(ctx context.Context, event domains.Event) (domains.NotificationParams, error) {
	if event.Params != nil {
		return *event.Params, nil
	}

	params, err := nj.repo.DefaultEventParams().Get(ctx, event.UserID)
	if err != nil {
		return domains.NotificationParams{}, fmt.Errorf("get default event params: %w", err)
	}

	return params, nil
}
