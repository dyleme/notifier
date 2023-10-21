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
)

type PeriodicEventRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) PeriodicEvents() service.PeriodicEventsRepository {
	return r.periodicEventRepository
}

func (p *PeriodicEventRepository) combineEventAndNotification(ev queries.PeriodicEvent, notif queries.PeriodicEventsNotification) domains.PeriodicEvent {
	return domains.PeriodicEvent{
		ID:                 int(ev.ID),
		Text:               ev.Text,
		Description:        pgxconv.String(ev.Description),
		UserID:             int(ev.UserID),
		Start:              pgxconv.TimeWithZone(ev.Start).Sub(time.Time{}),
		SmallestPeriod:     time.Duration(ev.SmallestPeriod) * time.Minute,
		BiggestPeriod:      time.Duration(ev.BiggestPeriod) * time.Minute,
		NotificationParams: ev.NotificationParams,
		Notification:       p.dtoPeriodicEventNotification(notif),
	}
}

func (p *PeriodicEventRepository) dtoPeriodicEventNotification(n queries.PeriodicEventsNotification) domains.PeriodicEventNotification {
	return domains.PeriodicEventNotification{
		ID:              int(n.ID),
		PeriodicEventID: int(n.PeriodicEventID),
		SendTime:        pgxconv.TimeWithZone(n.SendTime),
		Sended:          n.Sended,
		Done:            n.Done,
	}
}

func (p *PeriodicEventRepository) Add(ctx context.Context, event domains.PeriodicEvent, notif domains.PeriodicEventNotification) (domains.PeriodicEvent, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.AddPeriodicEvent(ctx, tx, queries.AddPeriodicEventParams{
		UserID:             int32(event.UserID),
		Text:               event.Text,
		Start:              pgxconv.Timestamptz(time.Time{}.Add(event.Start)),
		SmallestPeriod:     int32(event.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(event.BiggestPeriod / time.Minute),
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("add periodic event: %w", serverrors.NewRepositoryError(err))
	}

	n, err := p.q.AddPeriodicEventNotification(ctx, tx, queries.AddPeriodicEventNotificationParams{
		PeriodicEventID: ev.ID,
		SendTime:        pgxconv.Timestamptz(notif.SendTime),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("add periodic event notification[eventID=%v]: %w", ev.ID, serverrors.NewRepositoryError(err))
	}

	return p.combineEventAndNotification(ev, n), nil
}

func (p *PeriodicEventRepository) Get(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Get: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.GetPeriodicEvent(ctx, tx, queries.GetPeriodicEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	notif, err := p.q.CurrentPeriodicEventNotification(ctx, tx, ev.ID)
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("current periodic event notif[eventID=%v]: %w", ev.ID, serverrors.NewRepositoryError(err))
	}

	return p.combineEventAndNotification(ev, notif), nil
}

func (p *PeriodicEventRepository) Update(ctx context.Context, event service.UpdatePeriodicEventParams) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Update: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	ev, err := p.q.UpdatePeriodicEvent(ctx, tx, queries.UpdatePeriodicEventParams{
		ID:                 int32(event.ID),
		UserID:             int32(event.UserID),
		Start:              pgxconv.Timestamptz(time.Time{}.Add(event.Start)),
		Text:               event.Text,
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
		SmallestPeriod:     int32(event.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(event.BiggestPeriod / time.Minute),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	notif, err := p.q.CurrentPeriodicEventNotification(ctx, tx, ev.ID)
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf("current periodic event notification[eventID=%v]: %w", ev.ID, serverrors.NewRepositoryError(err))
	}

	return p.combineEventAndNotification(ev, notif), nil
}

func (p *PeriodicEventRepository) Delete(ctx context.Context, eventID, userID int) error {
	op := "PeriodicEventRepository.Delete: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	evs, err := p.q.DeletePeriodicEvent(ctx, tx, queries.DeletePeriodicEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("periodic event"))
	}

	return nil
}

func (p *PeriodicEventRepository) DeleteNotifications(ctx context.Context, eventID int) error {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	evs, err := p.q.DeletePeriodicEventNotifications(ctx, tx, int32(eventID))
	if err != nil {
		return fmt.Errorf("delete periodic notification: %w", serverrors.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf("evs == 0: %w", serverrors.NewNoDeletionsError("periodic event"))
	}

	return nil
}

func (p *PeriodicEventRepository) AddNotification(ctx context.Context, notification domains.PeriodicEventNotification) (domains.PeriodicEventNotification, error) {
	op := "PeriodicEventRepository.AddNotification: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	n, err := p.q.AddPeriodicEventNotification(ctx, tx, queries.AddPeriodicEventNotificationParams{
		PeriodicEventID: int32(notification.PeriodicEventID),
		SendTime:        pgxconv.Timestamptz(notification.SendTime),
	})
	if err != nil {
		return domains.PeriodicEventNotification{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dtoPeriodicEventNotification(n), nil
}

func (p *PeriodicEventRepository) MarkNotificationSend(ctx context.Context, notifID int) error {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	err := p.q.MarkPeriodicEventNotificationSended(ctx, tx, int32(notifID))
	if err != nil {
		return fmt.Errorf("mark notified: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func (p *PeriodicEventRepository) GetCurrentNotification(ctx context.Context, eventID int) (domains.PeriodicEventNotification, error) {
	op := "PeriodicEventRepository.GetCurrentNotification: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	notif, err := p.q.CurrentPeriodicEventNotification(ctx, tx, int32(eventID))
	if err != nil {
		return domains.PeriodicEventNotification{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return p.dtoPeriodicEventNotification(notif), nil
}

func (p *PeriodicEventRepository) DeleteNotification(ctx context.Context, notifID, eventID int) error {
	op := "PeriodicEventRepository.DeleteNotification: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	deletedNotifs, err := p.q.DeletePeriodicEventNotification(ctx, tx, queries.DeletePeriodicEventNotificationParams{
		ID:              int32(notifID),
		PeriodicEventID: int32(eventID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	if len(deletedNotifs) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("periodic event notification"))
	}

	return nil
}

func (p *PeriodicEventRepository) GetNearestNotificationSendTime(ctx context.Context) (time.Time, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	t, err := p.q.NearestPeriodicEventTime(ctx, tx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf("nearest notification send time: %w", serverrors.NewNotFoundError(err, "nearest time"))
		}

		return time.Time{}, fmt.Errorf("nearest notification send time: %w", serverrors.NewRepositoryError(err))
	}

	return pgxconv.TimeWithZone(t), nil
}

func (p *PeriodicEventRepository) ListNotificationsAtSendTime(ctx context.Context, sendTime time.Time) ([]domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.ListNotificationsAtSendTime: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	events, err := p.q.ListNearestPeriodicEvents(ctx, tx, pgxconv.Timestamptz(sendTime))
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return utils.DtoSlice(events, func(row queries.ListNearestPeriodicEventsRow) domains.PeriodicEvent {
		return p.combineEventAndNotification(queries.PeriodicEvent{
			ID:                 row.ID,
			CreatedAt:          row.CreatedAt,
			Text:               row.Text,
			Description:        row.Description,
			UserID:             row.UserID,
			Start:              row.Start,
			SmallestPeriod:     row.SmallestPeriod,
			BiggestPeriod:      row.BiggestPeriod,
			NotificationParams: row.NotificationParams,
		}, queries.PeriodicEventsNotification{
			ID:              row.ID_2,
			CreatedAt:       row.CreatedAt_2,
			PeriodicEventID: row.PeriodicEventID,
			SendTime:        row.SendTime,
			Sended:          row.Sended,
			Done:            row.Done,
		})
	}), nil
}

func (p *PeriodicEventRepository) ListFutureEvents(ctx context.Context, userID int, listParams service.ListParams) ([]domains.PeriodicEvent, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	events, err := p.q.ListPeriodicEventsWithNotifications(ctx, tx, queries.ListPeriodicEventsWithNotificationsParams{
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serverrors.NewNotFoundError(err, "periodic user events")
		}

		return nil, fmt.Errorf("list user periodic events: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoSlice(events, func(row queries.ListPeriodicEventsWithNotificationsRow) domains.PeriodicEvent {
		return p.combineEventAndNotification(queries.PeriodicEvent{
			ID:                 row.ID,
			CreatedAt:          row.CreatedAt,
			Text:               row.Text,
			Description:        row.Description,
			UserID:             row.UserID,
			Start:              row.Start,
			SmallestPeriod:     row.SmallestPeriod,
			BiggestPeriod:      row.BiggestPeriod,
			NotificationParams: row.NotificationParams,
		}, queries.PeriodicEventsNotification{
			ID:              row.ID_2,
			CreatedAt:       row.CreatedAt_2,
			PeriodicEventID: row.PeriodicEventID,
			SendTime:        row.SendTime,
			Sended:          row.Sended,
			Done:            row.Done,
		})
	}), nil
}

func (p *PeriodicEventRepository) MarkNotificationDone(ctx context.Context, eventID, userID int) error {
	tx := p.getter.DefaultTrOrDB(ctx, p.db)
	_, err := p.q.GetPeriodicEvent(ctx, tx, queries.GetPeriodicEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf("get periodic event: %w", serverrors.NewRepositoryError(err))
	}

	err = p.q.MarkPeriodicEventNotificationDone(ctx, tx, int32(eventID))
	if err != nil {
		return fmt.Errorf("mark periodic event notif done: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}
