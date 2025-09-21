package dailynotifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/log"
)

type Notifier interface {
	Notify(ctx context.Context, notif domain.Notification) error
}

type Repository interface {
	GetNextTime(ctx context.Context) (time.Time, error)
	DailyNotificationsUsers(ctx context.Context, now time.Time) ([]domain.User, error)
	ListDayEvents(ctx context.Context, userID, timeZoneOffset int) ([]domain.Sending, error)
	ListNotDoneEvents(ctx context.Context, userID int) ([]domain.Sending, error)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) (err error)
}

type DailyNotifier struct {
	repo     Repository
	tm       TxManager
	notifier Notifier
}

func New(repo Repository, tr TxManager) *DailyNotifier {
	return &DailyNotifier{
		repo:     repo,
		tm:       tr,
		notifier: nil,
	}
}

func (dn *DailyNotifier) SetNotifier(notifier Notifier) {
	dn.notifier = notifier
}

func (dn *DailyNotifier) GetNextTime(ctx context.Context) (time.Time, bool) {
	t, err := dn.repo.GetNextTime(ctx)
	if err != nil {
		var notFoundErr apperr.NotFoundError
		if !errors.As(err, &notFoundErr) {
			log.Ctx(ctx).Error("get next time error", log.Err(err))
		}

		return time.Time{}, false
	}

	return t, true
}

func (dn *DailyNotifier) Do(ctx context.Context, now time.Time) {
	err := dn.tm.Do(ctx, func(ctx context.Context) error {
		notifiedUsers, err := dn.repo.DailyNotificationsUsers(ctx, now)
		if err != nil {
			return fmt.Errorf("daily notifications users: %w", err)
		}

		for _, user := range notifiedUsers {
			events, err := dn.repo.ListDayEvents(ctx, user.ID, user.TimeZoneOffset)
			if err != nil {
				log.Ctx(ctx).Error("list events error", log.Err(err), slog.Time("run_time", now))
			}

			notDoneEvents, err := dn.repo.ListNotDoneEvents(ctx, user.ID)
			if err != nil {
				log.Ctx(ctx).Error("list not done events error", log.Err(err), slog.Time("run_time", now))
			}

			log.Ctx(ctx).Debug("user events", slog.Int("user_id", user.ID), slog.Any("events", events), slog.Any("not_done_events", notDoneEvents))

			err = dn.notifier.Notify(ctx, domain.Notification{})
			if err != nil {
				log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", now))
			}
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("notify error", log.Err(err), slog.Time("run_time", now))
	}
}
