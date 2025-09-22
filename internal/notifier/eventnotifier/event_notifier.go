package eventnotifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type Notifier interface {
	Notify(ctx context.Context, event domain.Event) error
}

type Service interface {
	ListNotSentEvents(ctx context.Context, till time.Time) ([]domain.Event, error)
	RescheduleSending(ctx context.Context, event domain.Event) error
	GetNearest(ctx context.Context) (time.Time, error)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
}

type Config struct {
	CheckTasksPeriod time.Duration
}

type EventNotifier struct {
	serv     Service
	notifier Notifier
	tm       TxManager
}

func New(tr TxManager) *EventNotifier {
	return &EventNotifier{
		serv:     nil,
		tm:       tr,
		notifier: nil,
	}
}

func (en *EventNotifier) SetNotifier(notifier Notifier) {
	en.notifier = notifier
}

func (en *EventNotifier) SetService(serv Service) {
	en.serv = serv
}

func (en *EventNotifier) GetNextTime(ctx context.Context) (time.Time, bool) {
	t, err := en.serv.GetNearest(ctx)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			log.Ctx(ctx).Debug("no events found")
		} else {
			log.Ctx(ctx).Error("get nearest event", log.Err(err))
		}

		return time.Time{}, false
	}

	return t, true
}

func (en *EventNotifier) Do(ctx context.Context, now time.Time) {
	err := en.tm.Do(ctx, func(ctx context.Context) error {
		events, err := en.serv.ListNotSentEvents(ctx, now)
		if err != nil {
			return fmt.Errorf("list not sended events: %w", err)
		}
		log.Ctx(ctx).Info("found not sended events", slog.Any("events", slice.Dto(events, func(e domain.Event) int { return e.SendingID })))

		for _, ev := range events {
			err = en.notifier.Notify(ctx, ev)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))

				continue
			}

			err = en.serv.RescheduleSending(ctx, ev)
			if err != nil {
				log.Ctx(ctx).Error("reschedule error", log.Err(err))
			}
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", now))
	}
}
