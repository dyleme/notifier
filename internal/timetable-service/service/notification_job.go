package service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/utils/dto"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

type Notifier interface {
	AddList(ctx context.Context, ns []domains.SendingNotification) error
	Add(ctx context.Context, n domains.SendingNotification) error
	Delete(ctx context.Context, id int) error
}

func (s *Service) RunJob(ctx context.Context) {
	ticker := time.NewTicker(s.checkTaskPeriod)
	for {
		select {
		case <-ticker.C:
			s.notify(ctx)
		case <-ctx.Done():
			ticker.Stop()

			return
		}
	}
}

func getNotifParams(ctx context.Context, r Repository, event *domains.Event) (domains.NotificationParams, error) {
	op := "getNotifParams: %w"
	if event.Notification.NotificationParams == nil {
		var err error
		userParam, err := r.DefaultNotificationParams().Get(ctx, event.UserID)
		if err != nil {
			return domains.NotificationParams{}, fmt.Errorf(op, err)
		}

		return userParam, nil
	}

	return *event.Notification.NotificationParams, nil
}

func mapNotifications(ctx context.Context, r Repository, events []domains.Event) ([]domains.SendingNotification, error) {
	op := "mapNotifications: %w"
	notifs, err := dto.ErrorContinueSlice(events, func(t domains.Event) (domains.SendingNotification, error) {
		notifParams, err := getNotifParams(ctx, r, &t)
		if err != nil {
			log.Ctx(ctx).Error("get_notif_params_error", log.Err(err))

			return domains.SendingNotification{}, err
		}

		return domains.SendingNotification{
			EventID:          t.ID,
			UserID:           t.UserID,
			Message:          t.Text,
			Description:      t.Description,
			NotificationTime: t.Start,
			Params:           notifParams,
		}, nil
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return notifs, nil
}

func (s *Service) notify(ctx context.Context) {
	op := "Service.notify: %w"
	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		Events, err := r.Events().GetNotNotified(ctx)
		if err != nil {
			return err //nolint:wrapcheck //wraping later
		}
		log.Ctx(ctx).Info("not_notified_from_database", "amount", len(Events))

		if len(Events) == 0 {
			return nil
		}

		wg, wgCtx := errgroup.WithContext(ctx)
		wg.SetLimit(1)

		wg.Go(func() error {
			notifs, err := mapNotifications(ctx, r, Events) //nolint:govet //need to shadow error
			if err != nil {
				return err
			}

			log.Ctx(ctx).Debug("add_notifications", "notifs", notifs)

			return s.notifier.AddList(wgCtx, notifs) //nolint:wrapcheck //wraping later
		})

		wg.Go(func() error {
			return r.Events().MarkNotified(wgCtx, dto.Slice(Events, func(t domains.Event) int { //nolint:wrapcheck //wraping later
				return t.ID
			}))
		})

		err = wg.Wait()
		if err != nil {
			return err //nolint:wrapcheck //wraping later
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("server error", log.Err(fmt.Errorf(op, err)))
	}
}
