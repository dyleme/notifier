package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries"
	"github.com/Dyleme/Notifier/internal/service/service"
	serverrors2 "github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
)

type PeriodicEventRepository struct {
	q *queries.Queries
}

func (r *Repository) PeriodicEvents() service.PeriodicEventsRepository {
	return &PeriodicEventRepository{q: r.q}
}

func (p *PeriodicEventRepository) dtoPeriodicEvent(ev queries.PeriodicEvent) domains.PeriodicEvent {
	return domains.PeriodicEvent{
		ID:                 int(ev.ID),
		Text:               ev.Text,
		Description:        pgxconv.String(ev.Description),
		UserID:             int(ev.UserID),
		Start:              p.appStart(pgxconv.TimeWithZone(ev.Start)),
		SmallestPeriod:     time.Duration(ev.SmallestPeriod),
		BiggestPeriod:      time.Duration(ev.BiggestPeriod),
		NotificationParams: ev.NotificationParams,
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

func (p *PeriodicEventRepository) Add(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Add: %w"
	ev, err := p.q.AddPeriodicEvent(ctx, queries.AddPeriodicEventParams{
		UserID:             int32(event.UserID),
		Text:               event.Text,
		Start:              pgxconv.Timestamptz(time.Time{}.Add(event.Start)),
		SmallestPeriod:     int32(event.SmallestPeriod),
		BiggestPeriod:      int32(event.BiggestPeriod),
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return p.dtoPeriodicEvent(ev), nil
}

func (p *PeriodicEventRepository) Get(ctx context.Context, eventID, userID int) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Get: %w"

	ev, err := p.q.GetPeriodicEvent(ctx, queries.GetPeriodicEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return p.dtoPeriodicEvent(ev), nil
}

func (p *PeriodicEventRepository) dbStart(duration time.Duration) time.Time {
	return time.Time{}.Add(duration)
}

func (p *PeriodicEventRepository) appStart(t time.Time) time.Duration {
	return time.Time{}.Sub(t)
}

func (p *PeriodicEventRepository) Update(ctx context.Context, event domains.PeriodicEvent) (domains.PeriodicEvent, error) {
	op := "PeriodicEventRepository.Update: %w"

	ev, err := p.q.UpdatePeriodicEvent(ctx, queries.UpdatePeriodicEventParams{
		ID:                 int32(event.ID),
		UserID:             int32(event.UserID),
		Start:              pgxconv.Timestamptz(p.dbStart(event.Start)),
		Text:               event.Text,
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
		SmallestPeriod:     int32(event.SmallestPeriod),
		BiggestPeriod:      int32(event.BiggestPeriod),
	})
	if err != nil {
		return domains.PeriodicEvent{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return p.dtoPeriodicEvent(ev), nil
}

func (p *PeriodicEventRepository) Delete(ctx context.Context, eventID, userID int) error {
	op := "PeriodicEventRepository.Delete: %w"

	evs, err := p.q.DeletePeriodicEvent(ctx, queries.DeletePeriodicEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, serverrors2.NewNoDeletionsError("periodic event"))
	}

	return nil
}

func (p *PeriodicEventRepository) AddNotification(ctx context.Context, notification domains.PeriodicEventNotification) (domains.PeriodicEventNotification, error) {
	op := "PeriodicEventRepository.AddNotification: %w"

	n, err := p.q.AddPeriodicEventNotification(ctx, queries.AddPeriodicEventNotificationParams{
		PeriodicEventID: int32(notification.PeriodicEventID),
		SendTime:        pgxconv.Timestamptz(notification.SendTime),
	})
	if err != nil {
		return domains.PeriodicEventNotification{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return p.dtoPeriodicEventNotification(n), nil
}

func (p *PeriodicEventRepository) UpdateNotification(ctx context.Context, notif domains.PeriodicEventNotification) error {
	op := "PeriodicEventRepository.UpdateNotification: %w"

	err := p.q.UpdatePeriodicEventNotification(ctx, queries.UpdatePeriodicEventNotificationParams{
		ID:              int32(notif.ID),
		PeriodicEventID: int32(notif.PeriodicEventID),
		Sended:          notif.Sended,
		SendTime:        pgxconv.Timestamptz(notif.SendTime),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return nil
}

func (p *PeriodicEventRepository) GetCurrentNotification(ctx context.Context, eventID int) (domains.PeriodicEventNotification, error) {
	op := "PeriodicEventRepository.GetCurrentNotification: %w"

	notif, err := p.q.CurrentPeriodicEventNotification(ctx, int32(eventID))
	if err != nil {
		return domains.PeriodicEventNotification{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return p.dtoPeriodicEventNotification(notif), nil
}

func (p *PeriodicEventRepository) DeleteNotification(ctx context.Context, notifID, eventID int) error {
	op := "PeriodicEventRepository.DeleteNotification: %w"

	deletedNotifs, err := p.q.DeletePeriodicEventNotification(ctx, queries.DeletePeriodicEventNotificationParams{
		ID:              int32(notifID),
		PeriodicEventID: int32(eventID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	if len(deletedNotifs) == 0 {
		return fmt.Errorf(op, serverrors2.NewNoDeletionsError("periodic event notification"))
	}

	return nil
}

func (p *PeriodicEventRepository) GetNearestNotificationSendTime(ctx context.Context) (time.Time, error) {
	op := "PeriodicEventRepository.GetNearestNotificationSendTime: %w"

	t, err := p.q.NearestPeriodicEventTime(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return pgxconv.TimeWithZone(t), nil
}

func (p *PeriodicEventRepository) ListNotificationsAtSendTime(ctx context.Context, sendTime time.Time) ([]domains.PeriodicEventWithNotification, error) {
	op := "PeriodicEventRepository.ListNotificationsAtSendTime: %w"

	events, err := p.q.ListNearestPeriodicEvents(ctx, pgxconv.Timestamptz(sendTime))
	if err != nil {
		return nil, fmt.Errorf(op, serverrors2.NewRepositoryError(err))
	}

	return utils.DtoSlice(events, func(row queries.ListNearestPeriodicEventsRow) domains.PeriodicEventWithNotification {
		return domains.PeriodicEventWithNotification{
			Event: p.dtoPeriodicEvent(queries.PeriodicEvent{
				ID:                 row.ID,
				CreatedAt:          row.CreatedAt,
				Text:               row.Text,
				Description:        row.Description,
				UserID:             row.UserID,
				Start:              row.Start,
				SmallestPeriod:     row.SmallestPeriod,
				BiggestPeriod:      row.BiggestPeriod,
				NotificationParams: row.NotificationParams,
			}),
			Notification: p.dtoPeriodicEventNotification(
				queries.PeriodicEventsNotification{
					ID:              row.ID_2,
					PeriodicEventID: row.PeriodicEventID,
					SendTime:        row.SendTime,
					Sended:          row.Sended,
					Done:            row.Done,
				},
			),
		}
	}), nil
}
