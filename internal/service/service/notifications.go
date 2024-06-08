package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

//go:generate mockgen -destination=mocks/notifications.go -package=mocks . NotificationsRepository
type NotificationsRepository interface { //nolint:interfacebloat // need so many interfaces
	Add(ctx context.Context, notification domains.Notification) (domains.Notification, error)
	List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams ListParams) ([]domains.Notification, error)
	Get(ctx context.Context, id int) (domains.Notification, error)
	GetLatest(ctx context.Context, eventdID int) (domains.Notification, error)
	Update(ctx context.Context, notification domains.Notification) error
	Delete(ctx context.Context, id int) error
	ListNotSended(ctx context.Context, till time.Time) ([]domains.Notification, error)
	GetNearest(ctx context.Context, till time.Time) (domains.Notification, error)
	MarkSended(ctx context.Context, ids []int) error
}

func (s *Service) ListNotifications(ctx context.Context, userID int, timeBorders timeborders.TimeBorders, listParams ListParams) ([]domains.Notification, error) {
	var notifications []domains.Notification
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error
		notifications, err = s.repo.Notifications().List(ctx, userID, timeBorders, listParams)
		if err != nil {
			return fmt.Errorf("notifications: list: %w", err)
		}

		return nil
	})
	if err != nil {
		logError(ctx, err)
		return nil, err
	}

	return notifications, nil
}
