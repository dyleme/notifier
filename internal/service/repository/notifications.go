package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

type NotificationsRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) Notifications() service.NotificationsRepository {
	return r.notificationsRepository
}

const UnkownTaskType = queries.TaskType("unknown_type")

func (*NotificationsRepository) repoTaskType(taskType domains.TaskType) (queries.TaskType, error) {
	switch taskType {
	case domains.PeriodicTaskType:
		return queries.TaskTypePeriodicTask, nil
	case domains.BasicTaskType:
		return queries.TaskTypeBasicTask, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (*NotificationsRepository) domainTaskType(taskType queries.TaskType) (domains.TaskType, error) {
	switch taskType {
	case queries.TaskTypePeriodicTask:
		return domains.PeriodicTaskType, nil
	case queries.TaskTypeBasicTask:
		return domains.BasicTaskType, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (n *NotificationsRepository) dto(notif queries.Notification) (domains.Notification, error) {
	taskType, err := n.domainTaskType(notif.TaskType)
	if err != nil {
		return domains.Notification{}, fmt.Errorf("domain task type: %w", err)
	}

	return domains.Notification{
		ID:          int(notif.ID),
		UserID:      int(notif.UserID),
		Text:        notif.Text,
		Description: pgxconv.String(notif.Description),
		TaskType:    taskType,
		TaskID:      int(notif.TaskID),
		Params:      notif.NotificationParams,
		SendTime:    pgxconv.TimeWithZone(notif.SendTime),
		Sended:      notif.Sended,
		Done:        notif.Done,
	}, nil
}

func (n *NotificationsRepository) Add(ctx context.Context, notification domains.Notification) (domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	taskType, err := n.repoTaskType(notification.TaskType)
	if err != nil {
		return domains.Notification{}, fmt.Errorf("repo task type: %w", err)
	}
	notif, err := n.q.AddNotification(ctx, tx, queries.AddNotificationParams{
		UserID:   int32(notification.UserID),
		Text:     notification.Text,
		TaskID:   int32(notification.TaskID),
		TaskType: taskType,
		SendTime: pgxconv.Timestamptz(notification.SendTime),
	})
	if err != nil {
		return domains.Notification{}, fmt.Errorf("add notification: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(notif)
}

func (n *NotificationsRepository) List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams service.ListParams) ([]domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	notifs, err := n.q.ListUserNotifications(ctx, tx, queries.ListUserNotificationsParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamptz(timeBorderes.From),
		ToTime:   pgxconv.Timestamptz(timeBorderes.To),
		Off:      int32(listParams.Offset),
		Lim:      int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list user notifications: %w", serverrors.NewRepositoryError(err))
	}

	domainNotifs, err := utils.DtoErrorSlice(notifs, n.dto)
	if err != nil {
		return nil, fmt.Errorf("list user notifications: %w", serverrors.NewRepositoryError(err))
	}

	return domainNotifs, nil
}

func (n *NotificationsRepository) Get(ctx context.Context, id int) (domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	notif, err := n.q.GetNotification(ctx, tx, int32(id))
	if err != nil {
		return domains.Notification{}, fmt.Errorf("get notification: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(notif)
}

func (n *NotificationsRepository) GetLatest(ctx context.Context, taskdID int) (domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	notif, err := n.q.GetLatestNotification(ctx, tx, int32(taskdID))
	if err != nil {
		return domains.Notification{}, fmt.Errorf("get latest notification: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(notif)
}

func (n *NotificationsRepository) Update(ctx context.Context, notification domains.Notification) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	_, err := n.q.UpdateNotification(ctx, tx, queries.UpdateNotificationParams{
		ID:       int32(notification.ID),
		Text:     notification.Text,
		SendTime: pgxconv.Timestamptz(notification.SendTime),
		Sended:   notification.Sended,
		Done:     notification.Done,
	})
	if err != nil {
		return fmt.Errorf("update notification: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func (n *NotificationsRepository) Delete(ctx context.Context, id int) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	ns, err := n.q.DeleteNotification(ctx, tx, int32(id))
	if err != nil {
		return fmt.Errorf("delete notification: %w", serverrors.NewRepositoryError(err))
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete notification: %w", serverrors.NewNoDeletionsError("notification"))
	}

	return nil
}

func (n *NotificationsRepository) ListNotSended(ctx context.Context, till time.Time) ([]domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	ns, err := n.q.ListNotSendedNotifications(ctx, tx, pgxconv.Timestamptz(till))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	domainNotifs, err := utils.DtoErrorSlice(ns, n.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return domainNotifs, nil
}

func (n *NotificationsRepository) GetNearest(ctx context.Context, till time.Time) (domains.Notification, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	notif, err := n.q.GetNearestNotification(ctx, tx, pgxconv.Timestamptz(till))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.Notification{}, fmt.Errorf("get nearest notification: %w", serverrors.NewNotFoundError(err, "notification"))
		}

		return domains.Notification{}, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(notif)
}

func (n *NotificationsRepository) MarkSended(ctx context.Context, ids []int) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	int32IDs := utils.DtoSlice(ids, func(i int) int32 { return int32(i) })
	err := n.q.MarkSendedNotifiatoins(ctx, tx, int32IDs)
	if err != nil {
		return fmt.Errorf("mark sended notifications: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}
