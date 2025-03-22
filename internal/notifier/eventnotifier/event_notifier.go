package eventnotifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type Notifier interface {
	Notify(ctx context.Context, notif domain.Notification) error
}

type Repository interface {
	Update(ctx context.Context, event domain.Event) error
	ListNotSended(ctx context.Context, till time.Time) ([]domain.Event, error)
	GetNearest(ctx context.Context) (time.Time, error)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) (err error)
}

type Config struct {
	CheckTasksPeriod time.Duration
}

type EventNotifier struct {
	repo     Repository
	notifier Notifier
	tm       TxManager
}

func New(repo Repository, tr TxManager) *EventNotifier {
	return &EventNotifier{
		repo:     repo,
		tm:       tr,
		notifier: nil,
	}
}

func (en *EventNotifier) SetNotifier(notifier Notifier) {
	en.notifier = notifier
}

func (en *EventNotifier) GetNextTime(ctx context.Context) (time.Time, bool) {
	t, err := en.repo.GetNearest(ctx)
	if err != nil {
		log.Ctx(ctx).Error("get nearest event", log.Err(err))

		return time.Time{}, false
	}

	return t, true
}

func (en *EventNotifier) Do(ctx context.Context, now time.Time) {
	err := en.tm.Do(ctx, func(ctx context.Context) error {
		events, err := en.repo.ListNotSended(ctx, now)
		if err != nil {
			return fmt.Errorf("list not sended events: %w", err)
		}
		log.Ctx(ctx).Info("found not sended events", slog.Any("events", slice.DtoSlice(events, func(n domain.Event) int { return n.ID })))

		for _, ev := range events {
			notification, err := ev.NewNotification()
			if err != nil {
				log.Ctx(ctx).Error("new sending event", log.Err(err))

				continue
			}
			err = en.notifier.Notify(ctx, notification)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))

				continue
			}

			ev = ev.Rescheule(now)

			err = en.repo.Update(ctx, ev)
			log.Ctx(ctx).Info("update event", slog.Any("event", ev))
			if err != nil {
				log.Ctx(ctx).Error("update event", log.Err(err), slog.Any("event", ev))
			}
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", now))
	}
}
