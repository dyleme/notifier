package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type NotificationParamsRepository interface {
	Get(ctx context.Context, userID int) (models.NotificationParams, error)
	Set(ctx context.Context, userID int, params models.NotificationParams) (models.NotificationParams, error)
}

func (s *Service) SetDefaultNotificationParams(ctx context.Context, params models.NotificationParams, userID int) (models.NotificationParams, error) {
	op := "Service.SetDefaultNotificationParams: %w"
	defParams, err := s.repo.DefaultNotificationParams().Set(ctx, userID, params)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return models.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetDefaultNotificationParams(ctx context.Context, userID int) (models.NotificationParams, error) {
	op := "Service.GetDefaultNotificationParams: %w"
	defParams, err := s.repo.DefaultNotificationParams().Get(ctx, userID)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return models.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetNotificationParams(ctx context.Context, timetableTaskID, userID int) (*models.NotificationParams, error) {
	op := "Service.GetNotificationParams: %w"
	var notifParams models.NotificationParams
	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		tt, err := repo.TimetableTasks().Get(ctx, timetableTaskID, userID)
		if err != nil {
			return err
		}

		if tt.Notification.Params != nil {
			notifParams = *tt.Notification.Params
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

func (s *Service) SetNotificationParams(ctx context.Context, timetableTaskID int, params models.NotificationParams, userID int) (models.NotificationParams, error) {
	op := "Service.SetNotificationParams: %w"
	updatedParams, err := s.repo.TimetableTasks().UpdateNotificationParams(ctx, timetableTaskID, userID, params)
	if err != nil {
		logError(ctx, fmt.Errorf(op, err))
		return models.NotificationParams{}, err
	}

	return updatedParams, nil
}
