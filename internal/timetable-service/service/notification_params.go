package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
)

type NotificationParamsRepository interface {
	Get(ctx context.Context, userID int) (domains.NotificationParams, error)
	Set(ctx context.Context, userID int, params domains.NotificationParams) (domains.NotificationParams, error)
}

func (s *Service) SetDefaultNotificationParams(ctx context.Context, params domains.NotificationParams, userID int) (domains.NotificationParams, error) {
	op := "Service.SetDefaultNotificationParams: %w"
	defParams, err := s.repo.DefaultNotificationParams().Set(ctx, userID, params)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetDefaultNotificationParams(ctx context.Context, userID int) (domains.NotificationParams, error) {
	op := "Service.GetDefaultNotificationParams: %w"
	defParams, err := s.repo.DefaultNotificationParams().Get(ctx, userID)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetNotificationParams(ctx context.Context, eventID, userID int) (*domains.NotificationParams, error) {
	op := "Service.GetNotificationParams: %w"
	var notifParams domains.NotificationParams
	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		tt, err := repo.Events().Get(ctx, eventID, userID)
		if err != nil {
			return err
		}

		if tt.Notification.NotificationParams != nil {
			notifParams = *tt.Notification.NotificationParams

			return nil
		}

		defaultParams, err := repo.DefaultNotificationParams().Get(ctx, userID)
		if err != nil {
			var notFoundErr serverrors.NotFoundError
			if errors.As(err, &notFoundErr) {
				return err
			}

			return err
		}

		notifParams = defaultParams

		return nil
	})
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))

		return nil, err
	}

	return &notifParams, nil
}

func (s *Service) SetNotificationParams(ctx context.Context, eventID int, params domains.NotificationParams, userID int) (domains.NotificationParams, error) {
	op := "Service.SetNotificationParams: %w"
	updatedParams, err := s.repo.Events().UpdateNotificationParams(ctx, eventID, userID, params)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))

		return domains.NotificationParams{}, err
	}

	return updatedParams, nil
}
