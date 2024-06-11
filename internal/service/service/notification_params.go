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

func (s *Service) SetDefaultEventParams(ctx context.Context, params domains.NotificationParams, userID int) (domains.NotificationParams, error) {
	defParams, err := s.repo.DefaultEventParams().Set(ctx, userID, params)
	if err != nil {
		err = fmt.Errorf("set deafault event params: %w", err)
		logError(ctx, err)

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}

func (s *Service) GetDefaultEventParams(ctx context.Context, userID int) (domains.NotificationParams, error) {
	defParams, err := s.repo.DefaultEventParams().Get(ctx, userID)
	if err != nil {
		err = fmt.Errorf("get deafault event params: %w", err)
		logError(ctx, err)

		return domains.NotificationParams{}, err
	}

	return defParams, nil
}
