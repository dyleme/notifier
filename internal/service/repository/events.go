package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/database/sqlconv"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type EventsRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewEventsRepository(getter *txmanager.Getter) *EventsRepository {
	return &EventsRepository{
		getter: getter,
		q:      &goqueries.Queries{},
	}
}

func (er *EventsRepository) dto(dbEv goqueries.Event) (domain.Event, error) {
	event := domain.Event{
		ID:              int(dbEv.ID),
		CreatedAt:       dbEv.CreatedAt,
		TaskID:          int(dbEv.TaskID),
		Done:            sqlconv.ToBool(dbEv.Done),
		OriginalSending: dbEv.OriginalSending,
		NextSending:     dbEv.NextSending,
	}

	return event, nil
}

func (er *EventsRepository) Add(ctx context.Context, event domain.Event) error {
	tx := er.getter.GetTx(ctx)

	_, err := er.q.AddEvent(ctx, tx, goqueries.AddEventParams{
		TaskID:          int64(event.TaskID),
		Done:            sqlconv.BoolToInt(event.Done),
		OriginalSending: event.OriginalSending,
		NextSending:     event.NextSending,
	})
	if err != nil {
		return fmt.Errorf("add event: %w", err)
	}

	return nil
}

func (er *EventsRepository) List(ctx context.Context, userID int, params service.ListEventsFilterParams) ([]domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	rowsEvents, err := er.q.ListUserEvents(ctx, tx, goqueries.ListUserEventsParams{
		UserID:   int64(userID),
		FromTime: params.TimeBorders.From,
		ToTime:   params.TimeBorders.To,
		Offset:   int64(params.ListParams.Offset),
		Limit:    int64(params.ListParams.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", err)
	}

	events, err := slice.DtoError(rowsEvents, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", err)
	}

	return events, nil
}

func (er *EventsRepository) Get(ctx context.Context, id int) (domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	event, err := er.q.GetEvent(ctx, tx, int64(id))
	if err != nil {
		return domain.Event{}, fmt.Errorf("get event: %w", err)
	}

	return er.dto(event)
}

func (er *EventsRepository) GetLatest(ctx context.Context, taskdID int) (domain.Event, error) {
	tx := er.getter.GetTx(ctx)
	event, err := er.q.GetLatestEvent(ctx, tx, int64(taskdID))
	if err != nil {
		return domain.Event{}, fmt.Errorf("get latest event: %w", err)
	}

	return er.dto(event)
}

func (er *EventsRepository) Update(ctx context.Context, event domain.Event) error {
	tx := er.getter.GetTx(ctx)
	_, err := er.q.UpdateEvent(ctx, tx, goqueries.UpdateEventParams{
		NextSending:     event.NextSending,
		OriginalSending: event.OriginalSending,
		Done:            sqlconv.BoolToInt(event.Done),
		ID:              int64(event.ID),
	})
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	return nil
}

func (er *EventsRepository) Delete(ctx context.Context, id int) error {
	tx := er.getter.GetTx(ctx)

	ns, err := er.q.DeleteEvent(ctx, tx, int64(id))
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete event: %w", apperr.ErrNotFound)
	}

	return nil
}

func (er *EventsRepository) ListNotSended(ctx context.Context, till time.Time) ([]domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	dbEvents, err := er.q.ListNotSendedEvents(ctx, tx, till)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	events, err := slice.DtoError(dbEvents, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return events, nil
}

func (er *EventsRepository) GetNearest(ctx context.Context) (time.Time, error) {
	tx := er.getter.GetTx(ctx)

	t, err := er.q.GetNearestEventTime(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, fmt.Errorf("get nearest event: %w", apperr.ErrNotFound)
		}

		return time.Time{}, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return t, nil
}
