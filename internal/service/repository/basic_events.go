package repository

import (
	"context"
	"errors"
	"fmt"

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

type EventRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) Events() service.BasicEventRepository {
	return r.eventsRepository
}

func dtoEvent(be queries.BasicEvent) (domains.BasicEvent, error) {
	return domains.BasicEvent{
		ID:                 int(be.ID),
		UserID:             int(be.UserID),
		Text:               be.Text,
		Description:        pgxconv.String(be.Description),
		Start:              pgxconv.TimeWithZone(be.Start),
		NotificationParams: be.NotificationParams,
	}, nil
}

func (er *EventRepository) Add(ctx context.Context, ev domains.BasicEvent) (domains.BasicEvent, error) {
	op := "add timetable task: %w"

	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	addedEvent, err := er.q.AddBasicEvent(ctx, tx, queries.AddBasicEventParams{
		UserID:             int32(ev.UserID),
		Text:               ev.Text,
		Start:              pgxconv.Timestamptz(ev.Start),
		Description:        pgxconv.Text(ev.Description),
		NotificationParams: ev.NotificationParams,
	})
	if err != nil {
		return domains.BasicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoEvent(addedEvent)
}

func (er *EventRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domains.BasicEvent, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} %%w", userID)
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	tt, err := er.q.ListBasicEvents(ctx, tx, queries.ListBasicEventsParams{
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

func (er *EventRepository) Delete(ctx context.Context, eventID int) error {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	deletedEvents, err := er.q.DeleteBasicEvent(ctx, tx, int32(eventID))
	if err != nil {
		return fmt.Errorf("delete basic event: %w", serverrors.NewRepositoryError(err))
	}
	if len(deletedEvents) == 0 {
		return serverrors.NewNoDeletionsError("events")
	}

	return nil
}

func (er *EventRepository) Get(ctx context.Context, eventID int) (domains.BasicEvent, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	tt, err := er.q.GetBasicEvent(ctx, tx, int32(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.BasicEvent{}, fmt.Errorf("get basic event: %w", serverrors.NewNotFoundError(err, "basic event"))
		}

		return domains.BasicEvent{}, fmt.Errorf("get basic event: %w", serverrors.NewRepositoryError(err))
	}

	return dtoEvent(tt)
}

func (er *EventRepository) Update(ctx context.Context, event domains.BasicEvent) (domains.BasicEvent, error) {
	op := "update timetable task: %w"
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	updatedTT, err := er.q.UpdateBasicEvent(ctx, tx, queries.UpdateBasicEventParams{
		Start:              pgxconv.Timestamptz(event.Start),
		Text:               event.Text,
		Description:        pgxconv.Text(event.Description),
		NotificationParams: event.NotificationParams,
		ID:                 int32(event.ID),
		UserID:             int32(event.UserID),
	})
	if err != nil {
		return domains.BasicEvent{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoEvent(updatedTT)
}
