package repository

import (
	"context"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type NotificationRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewNotificationRepository(getter *txmanager.Getter) *NotificationRepository {
	return &NotificationRepository{
		getter: getter,
		q:      &goqueries.Queries{},
	}
}

func (r *NotificationRepository) ListNotSent(
	ctx context.Context, till time.Time,
) ([]domain.Notification, error) {
	tx := r.getter.GetTx(ctx)
	dbNotifications, err := r.q.ListNotSentNotifications(ctx, tx, till)
	if err != nil {
		return nil, err
	}

	notifs := slice.Dto(dbNotifications,
		func(r goqueries.ListNotSentNotificationsRow) domain.Notification {
			return domain.Notification{
				EventID:            int(r.EventID),
				SendTime:           r.NextSending,
				Message:            r.Text,
				TgID:               int(r.TgID),
				NotificationPeriod: time.Duration(r.NotificationRetryPeriodS) * time.Second,
			}
		},
	)

	return notifs, nil
}
