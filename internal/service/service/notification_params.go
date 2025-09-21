package service

import (
	"context"

	"github.com/Dyleme/Notifier/internal/domain"
)

//go:generate mockgen -destination=mocks/notification_params_mocks.go -package=mocks . DefaultNotificationParamsRepository
type DefaultNotificationParamsRepository interface {
	Set(ctx context.Context, userID int, params domain.NotificationParams) (domain.NotificationParams, error)
	Get(ctx context.Context, userID int) (domain.NotificationParams, error)
}
