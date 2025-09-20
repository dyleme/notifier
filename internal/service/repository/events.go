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
	"github.com/Dyleme/Notifier/pkg/database/pgxconv"
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
	notifParams, err := parseNotificationParams(dbEv.NotificationParams)
	if err != nil {
		return domain.Event{}, fmt.Errorf("parse notification params: %w", err)
	}
	event := domain.Event{
		ID:                 int(dbEv.ID),
		UserID:             int(dbEv.UserID),
		Text:               dbEv.Text,
		Description:        dbEv.Description.String,
		TaskType:           domain.TaskType(dbEv.TaskType),
		TaskID:             int(dbEv.TaskID),
		NotificationParams: notifParams,
		NextSend:           dbEv.NextSend,
		FirstSend:          dbEv.FirstSend,
		Done:               sqlconv.ToBool(dbEv.Done),
		Tags:               nil,
		Notify:             sqlconv.ToBool(dbEv.Notify),
	}

	return event, nil
}

func (er *EventsRepository) dtoWithTags(dbEv goqueries.Event, dbTags []goqueries.Tag) (domain.Event, error) {
	notifParams, err := parseNotificationParams(dbEv.NotificationParams)
	if err != nil {
		return domain.Event{}, fmt.Errorf("parse notification params: %w", err)
	}
	event := domain.Event{
		ID:                 int(dbEv.ID),
		UserID:             int(dbEv.UserID),
		Text:               dbEv.Text,
		Description:        dbEv.Description.String,
		TaskType:           domain.TaskType(dbEv.TaskType),
		TaskID:             int(dbEv.TaskID),
		NotificationParams: notifParams,
		NextSend:           dbEv.NextSend,
		FirstSend:          dbEv.FirstSend,
		Done:               sqlconv.ToBool(dbEv.Done),
		Tags:               slice.Dto(dbTags, dtoTag),
		Notify:             sqlconv.ToBool(dbEv.Notify),
	}

	return event, nil
}

func (er *EventsRepository) Add(ctx context.Context, event domain.Event) (domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	ev, err := er.q.AddEvent(ctx, tx, goqueries.AddEventParams{
		UserID:             int64(event.UserID),
		Text:               event.Text,
		TaskID:             int64(event.TaskID),
		TaskType:           string(event.TaskType),
		NextSend:           event.NextSend,
		NotificationParams: event.NotificationParams.JSON(),
	})
	if err != nil {
		return domain.Event{}, fmt.Errorf("add event: %w", err)
	}

	for _, t := range event.Tags {
		err = er.q.AddTagsToSmth(ctx, tx,
			goqueries.AddTagsToSmthParams{
				SmthID: ev.ID,
				TagID:  int64(t.ID),
				UserID: int64(event.UserID),
			},
		)
	}
	if err != nil {
		return domain.Event{}, fmt.Errorf("add tags to smth: %w", err)
	}

	return er.Get(ctx, int(ev.ID))
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list user events: %w", err)
	}

	tasksIDs := slice.Dto(rowsEvents, func(t goqueries.ListUserEventsRow) int64 { return t.Event.ID })

	rows, err := listTagsForSmths(ctx, tx, er.q, tasksIDs)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("list tags for smths: %w", err)
		}
	}

	events := make([]domain.Event, 0, len(rowsEvents))
	for _, ev := range rowsEvents {
		var tags []goqueries.Tag
		for _, row := range rows {
			if row.SmthID == ev.Event.ID {
				tags = append(tags, row.Tag)
			}
		}
		event, err := er.dtoWithTags(ev.Event, tags)
		if err != nil {
			return nil, fmt.Errorf("event dto: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (er *EventsRepository) Get(ctx context.Context, id int) (domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	event, err := er.q.GetEvent(ctx, tx, int64(id))
	if err != nil {
		return domain.Event{}, fmt.Errorf("get event: %w", err)
	}

	tags, err := er.q.ListTagsForSmth(ctx, tx, event.ID)
	if err != nil {
		return domain.Event{}, fmt.Errorf("list tags for smth: %w", err)
	}

	return er.dtoWithTags(event, tags)
}

func (er *EventsRepository) GetLatest(ctx context.Context, taskdID int, taskType domain.TaskType) (domain.Event, error) {
	tx := er.getter.GetTx(ctx)

	event, err := er.q.GetLatestEvent(ctx, tx, goqueries.GetLatestEventParams{
		TaskID:   int64(taskdID),
		TaskType: string(taskType),
	})
	if err != nil {
		return domain.Event{}, fmt.Errorf("get latest event: %w", err)
	}

	tags, err := er.q.ListTagsForSmth(ctx, tx, event.ID)
	if err != nil {
		return domain.Event{}, fmt.Errorf("list tags for smth: %w", err)
	}

	return er.dtoWithTags(event, tags)
}

func (er *EventsRepository) Update(ctx context.Context, event domain.Event) error {
	tx := er.getter.GetTx(ctx)

	_, err := er.q.UpdateEvent(ctx, tx, goqueries.UpdateEventParams{
		Text:      event.Text,
		NextSend:  event.FirstSend,
		FirstSend: event.FirstSend,
		Done:      sqlconv.BoolToInt(event.Done),
		ID:        int64(event.ID),
	})
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	err = syncTags(ctx, tx, er.q, event.UserID, event.ID, event.Tags)
	if err != nil {
		return fmt.Errorf("sync tags: %w", err)
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

	events, err := er.q.ListNotSendedEvents(ctx, tx, till)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	domainEvents, err := slice.DtoError(events, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", err)
	}

	return domainEvents, nil
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

func (er *EventsRepository) ListDayEvents(ctx context.Context, userID, timeZoneOffset int) ([]domain.Event, error) {
	tx := er.getter.GetTx(ctx)
	events, err := er.q.ListUserDailyEvents(ctx, tx, goqueries.ListUserDailyEventsParams{
		UserID:     int64(userID),
		TimeOffset: pgxconv.Timestamptz(time.Time{}.Add(time.Duration(timeZoneOffset) * time.Hour)),
	})
	if err != nil {
		return nil, fmt.Errorf("list user daily events: %w", err)
	}

	return slice.DtoError(events, er.dto)
}

func (er *EventsRepository) ListNotDoneEvents(ctx context.Context, userID int) ([]domain.Event, error) {
	tx := er.getter.GetTx(ctx)
	events, err := er.q.ListNotDoneEvents(ctx, tx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("list not done events: %w", err)
	}

	return slice.DtoError(events, er.dto)
}
