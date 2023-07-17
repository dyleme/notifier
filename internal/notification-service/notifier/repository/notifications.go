package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Dyleme/Notifier/internal/notification-service/notifier/notification"
	"github.com/Dyleme/Notifier/internal/notification-service/notifier/repository/queries"
)

func (r *Repository) GetNewNotifications(ctx context.Context) ([]notification.Notification, error) {
	var (
		notifs []queries.Notification
		err    error
	)

	err = r.inTx(ctx, defaulTxOpts, func(q *queries.Queries) error {
		timeNow := time.Now()
		notifs, err = q.FetchNewNotifications(ctx, timeNow)
		if err != nil {
			return err
		}

		err = q.MarkSendedNotifications(ctx, timeNow)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dtoNotifications(notifs), nil
}

func dtoNotifications(notifs []queries.Notification) []notification.Notification {
	ns := make([]notification.Notification, 0, len(notifs))

	for _, n := range notifs {
		ns = append(ns, dtoNotification(n))
	}

	return ns
}

func dtoNotification(n queries.Notification) notification.Notification {
	return notification.Notification{
		ID:               int(n.ID),
		UserID:           int(n.UserID),
		Message:          n.Message,
		NotificationTime: n.NotificationTime,
		Destinations:     []string{},
		TaskID:           int(n.TaskID),
	}
}

func (r *Repository) AddNotification(ctx context.Context, n notification.Notification) error {
	var rm json.RawMessage
	err := rm.UnmarshalJSON([]byte(`{"notify":true}`))
	if err != nil {
		return err
	}
	return r.q.AddNotification(ctx, queries.AddNotificationParams{
		UserID:           int32(n.UserID),
		Message:          n.Message,
		NotificationTime: n.NotificationTime,
		Destination:      rm,
		TaskID:           int32(n.TaskID),
	})
}

func (r *Repository) GetNotification(ctx context.Context, id int) (notification.Notification, error) {
	n, err := r.q.GetNotification(ctx, int32(id))
	if err != nil {
		return notification.Notification{}, err
	}

	return dtoNotification(n), nil
}

func (r *Repository) GetFutureUserNotifications(ctx context.Context, userID int) ([]notification.Notification, error) {
	ns, err := r.q.GetFutureUserNotifications(ctx, queries.GetFutureUserNotificationsParams{
		FromTime: time.Now(),
		UserID:   int32(userID),
	})
	if err != nil {
		return nil, err
	}

	return dtoNotifications(ns), nil
}

func (r *Repository) DeleteNotification(ctx context.Context, id int) error {
	err := r.q.DeleteNotification(ctx, int32(id))
	if err != nil {
		return err
	}

	return nil
}
