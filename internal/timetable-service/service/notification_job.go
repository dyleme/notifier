package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/utils/dto"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type Notifier interface {
	Add(ctx context.Context, ns []models.SendingNotification) error
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

func getNotifParams(ctx context.Context, r Repository, t *models.TimetableTask, defaultParams map[int]models.NotificationParams) (models.NotificationParams, error) {
	op := "getNotifParams: %w"
	if t.Notification.Params == nil {
		userParam, ok := defaultParams[t.UserID]
		if !ok {
			var err error
			userParam, err = r.DefaultNotificationParams().Get(ctx, t.UserID)
			if err != nil {
				var notFoundErr serverrors.NotFoundError
				if errors.As(err, &notFoundErr) {
					return models.NotificationParams{}, fmt.Errorf(op, err)
				}
				return models.NotificationParams{}, fmt.Errorf(op, err)
			}
			defaultParams[t.UserID] = userParam
		}
		t.Notification.Params = &userParam
		return userParam, nil
	}
	return *t.Notification.Params, nil
}

func mapNotifications(ctx context.Context, r Repository, timetableTasks []models.TimetableTask) ([]models.SendingNotification, error) {
	op := "mapNotifications: %w"
	defaultParams := make(map[int]models.NotificationParams)

	notifs, err := dto.ErrorContinueSlice(timetableTasks, func(t models.TimetableTask) (models.SendingNotification, error) {
		notifParams, err := getNotifParams(ctx, r, &t, defaultParams)
		if err != nil {
			log.Ctx(ctx).Error("get_notif_params_error", log.Err(err))
			return models.SendingNotification{}, err
		}
		return models.SendingNotification{
			TaskID:           t.TaskID,
			TimetableTaskID:  t.ID,
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
		timetableTasks, err := r.TimetableTasks().GetNotNotified(ctx)
		if err != nil {
			return err
		}
		log.Ctx(ctx).Info("not_notified_from_database", "amount", len(timetableTasks))

		if len(timetableTasks) == 0 {
			return nil
		}

		wg, wgCtx := errgroup.WithContext(ctx)
		wg.SetLimit(1)

		wg.Go(func() error {
			notifs, err := mapNotifications(ctx, r, timetableTasks) //nolint:govet //need to shadow error
			if err != nil {
				return err
			}

			log.Ctx(ctx).Debug("add_notifications", "notifs", notifs)
			return s.notifier.Add(wgCtx, notifs)
		})

		wg.Go(func() error {
			return r.TimetableTasks().MarkNotified(wgCtx, dto.Slice(timetableTasks, func(t models.TimetableTask) int {
				return t.ID
			}))
		})

		err = wg.Wait()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error("server error", log.Err(fmt.Errorf(op, err)))
	}
}
