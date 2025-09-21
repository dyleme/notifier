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
	Notify(ctx context.Context, notif domain.Notification) error
}

type Repository interface {
	ListNotSent(ctx context.Context, till time.Time) ([]domain.Notification, error)
	Reschedule(ctx context.Context, eventID int, nextSendingTime time.Time) error
	GetNearest(ctx context.Context) (time.Time, error)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
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
		notifications, err := en.repo.ListNotSent(ctx, now)
		if err != nil {
			return fmt.Errorf("list not sended events: %w", err)
		}
		log.Ctx(ctx).Info("found not sended events", slog.Any("events", slice.Dto(notifications, func(n domain.Notification) int { return n.EventID })))

		for _, notif := range notifications {
			err = en.notifier.Notify(ctx, notif)
			if err != nil {
				log.Ctx(ctx).Error("notifier error", log.Err(err))

				continue
			}

			err = en.repo.Reschedule(ctx, notif.EventID, now.Add(notif.NotificationPeriod))
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
