package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
	"github.com/Dyleme/Notifier/internal/lib/utils"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type EventRepository struct {
	q *queries.Queries
}

func (r *Repository) Events() service.EventRepository {
	return &EventRepository{q: r.q}
}

func dtoEvent(ev queries.Event) (domains.Event, error) {
	return domains.Event{
		ID:                 int(ev.ID),
		UserID:             int(ev.UserID),
		Text:               ev.Text,
		Description:        pgxconv.String(ev.Description),
		Start:              pgxconv.TimeWithZone(ev.Start),
		Done:               ev.Done,
		SendTime:           pgxconv.TimeWithZone(ev.Start),
		Sended:             ev.Sended,
		NotificationParams: ev.NotificationParams,
	}, nil
}

func (er *EventRepository) Add(ctx context.Context, ev domains.Event) (domains.Event, error) {
	op := "add timetable task: %w"
	addedEvent, err := er.q.AddEvent(ctx, queries.AddEventParams{
		UserID:             int32(ev.UserID),
		Text:               ev.Text,
		Start:              pgxconv.Timestamptz(ev.Start),
		Description:        pgxconv.Text(ev.Description),
		Done:               ev.Done,
		NotificationParams: ev.NotificationParams,
		SendTime:           pgxconv.Timestamptz(ev.Start),
		Sended:             false,
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoEvent(addedEvent)
}

func (er *EventRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.Event, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} %%w", userID)
	tt, err := er.q.ListEvents(ctx, queries.ListEventsParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	events, err := utils.DtoErrorSlice(tt, dtoEvent)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return events, nil
}

func (er *EventRepository) Delete(ctx context.Context, eventID, userID int) error {
	op := fmt.Sprintf("delete timetable tasks eventID{%v} userID{%v} %%w", eventID, userID)
	deletedEvents, err := er.q.DeleteEvent(ctx, queries.DeleteEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if len(deletedEvents) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("events"))
	}

	return nil
}

func (er *EventRepository) ListInPeriod(ctx context.Context, userID int, from, to time.Time, params service.ListParams) ([]domains.Event, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} from{%v} to{%v}: %%w", userID, from, to)
	tts, err := er.q.GetEventsInPeriod(ctx, queries.GetEventsInPeriodParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamptz(from),
		ToTime:   pgxconv.Timestamptz(to),
		Off:      int32(params.Offset),
		Lim:      int32(params.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	events, err := utils.DtoErrorSlice(tts, dtoEvent)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return events, nil
}

func (er *EventRepository) Get(ctx context.Context, eventID, userID int) (domains.Event, error) {
	op := fmt.Sprintf("EventRepository.Get timetable tasks eventID{%v} userID{%v} %%w", eventID, userID)
	tt, err := er.q.GetEvent(ctx, queries.GetEventParams{
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.Event{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "timetable task"))
		}

		return domains.Event{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoEvent(tt)
}

func (er *EventRepository) Update(ctx context.Context, event domains.Event) (domains.Event, error) {
	op := "update timetable task: %w"
	updatedTT, err := er.q.UpdateEvent(ctx, queries.UpdateEventParams{
		Start:              pgxconv.Timestamptz(event.Start),
		Text:               event.Text,
		Description:        pgxconv.Text(event.Description),
		Done:               event.Done,
		NotificationParams: event.NotificationParams,
		ID:                 int32(event.ID),
		UserID:             int32(event.UserID),
		Sended:             event.Sended,
		SendTime:           pgxconv.Timestamptz(event.SendTime),
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoEvent(updatedTT)
}

func (er *EventRepository) GetNearestEventSendTime(ctx context.Context) (time.Time, error) {
	op := "EventRepository.GetNearestEventSendTime: %w"

	nearestTime, err := er.q.NearestEventTime(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "nearest time"))
		}

		return time.Time{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return pgxconv.TimeWithZone(nearestTime), nil
}

func (er *EventRepository) ListEventsBefore(ctx context.Context, sendTime time.Time) ([]domains.Event, error) {
	op := "EventRepository.ListEventsBefore: %w"

	events, err := er.q.ListNearestEvents(ctx, pgxconv.Timestamptz(sendTime))
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	notNotified, err := utils.DtoErrorSlice(events, dtoEvent)
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return notNotified, nil
}

func (er *EventRepository) MarkNotified(ctx context.Context, eventID int) error {
	op := "EventRepository.MarkNotified: %w"
	err := er.q.MarkSendedNotificationEvent(ctx, int32(eventID))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (er *EventRepository) UpdateNotificationParams(ctx context.Context, eventID, userID int, params domains.NotificationParams) (domains.NotificationParams, error) {
	op := "EventRepository.UpdateNotificationParams: %w"

	p, err := er.q.UpdateNotificationParams(ctx, queries.UpdateNotificationParamsParams{
		Params: &params,
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return domains.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	if p == nil {
		return domains.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(fmt.Errorf("params are nil after update")))
	}

	return *p, nil
}

func (er *EventRepository) Delay(ctx context.Context, eventID, userID int, till time.Time) error {
	op := "EventRepository.Delay: %w"
	err := er.q.DelayEvent(ctx, queries.DelayEventParams{
		Till:   pgxconv.Timestamp(till),
		ID:     int32(eventID),
		UserID: int32(userID),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}
