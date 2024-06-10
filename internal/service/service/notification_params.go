package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domains"
)

//go:generate mockgen -destination=mocks/notification_params_mocks.go -package=mocks . NotificationParamsRepository
type NotificationParamsRepository interface {
	Set(ctx context.Context, userID int, params domains.NotificationParams) (domains.NotificationParams, error)
	Get(ctx context.Context, userID int) (domains.NotificationParams, error)
}

func (s *Service) SetDefaultNotificationParams(ctx context.Context, params domains.NotificationParams, userID int) (domains.NotificationParams, error) {
	defParams, err := s.repo.DefaultNotificationParams().Set(ctx, userID, params)
	if err != nil {
		err = fmt.Errorf("set deafault notification params: %w", err)
		logError(ctx, err)

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetDefaultNotificationParams(ctx context.Context, userID int) (domains.NotificationParams, error) {
	defParams, err := s.repo.DefaultNotificationParams().Get(ctx, userID)
	if err != nil {
		err = fmt.Errorf("get deafault notification params: %w", err)
		logError(ctx, err)

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}
