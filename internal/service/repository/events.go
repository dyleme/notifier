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
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/sql/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils"
	"github.com/Dyleme/Notifier/pkg/utils/timeborders"
)

type EventsRepository struct {
	q      *goqueries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewEventsRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *EventsRepository {
	return &EventsRepository{
		q:      goqueries.New(),
		db:     db,
		getter: getter,
	}
}

const UnkownTaskType = goqueries.TaskType("unknown_type")

func (*EventsRepository) repoTaskType(taskType domains.TaskType) (goqueries.TaskType, error) {
	switch taskType {
	case domains.PeriodicTaskType:
		return goqueries.TaskTypePeriodicTask, nil
	case domains.BasicTaskType:
		return goqueries.TaskTypeBasicTask, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (*EventsRepository) domainTaskType(taskType goqueries.TaskType) (domains.TaskType, error) {
	switch taskType {
	case goqueries.TaskTypePeriodicTask:
		return domains.PeriodicTaskType, nil
	case goqueries.TaskTypeBasicTask:
		return domains.BasicTaskType, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (er *EventsRepository) dto(event goqueries.Event) (domains.Event, error) {
	taskType, err := er.domainTaskType(event.TaskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("domain task type: %w", err)
	}

	return domains.Event{
		ID:                 int(event.ID),
		UserID:             int(event.UserID),
		Text:               event.Text,
		Description:        pgxconv.String(event.Description),
		TaskType:           taskType,
		TaskID:             int(event.TaskID),
		NotificationParams: event.NotificationParams,
		LastSendedTime:     pgxconv.TimeWithZone(event.LastSendedTime),
		NextSendTime:       pgxconv.TimeWithZone(event.NextSendTime),
		FirstSendTime:      pgxconv.TimeWithZone(event.FirstSendTime),
		Done:               event.Done,
	}, nil
}

func (er *EventsRepository) Add(ctx context.Context, event domains.Event) (domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	taskType, err := er.repoTaskType(event.TaskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("repo task type: %w", err)
	}
	ev, err := er.q.AddEvent(ctx, tx, goqueries.AddEventParams{
		UserID:             int32(event.UserID),
		Text:               event.Text,
		TaskID:             int32(event.TaskID),
		TaskType:           taskType,
		NextSendTime:       pgxconv.Timestamptz(event.NextSendTime),
		NotificationParams: event.NotificationParams,
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf("add event: %w", serverrors.NewRepositoryError(err))
	}

	return er.dto(ev)
}

func (er *EventsRepository) List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams service.ListParams) ([]domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	notifs, err := er.q.ListUserEvents(ctx, tx, goqueries.ListUserEventsParams{
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

		return nil, fmt.Errorf("list user events: %w", serverrors.NewRepositoryError(err))
	}

	domainNotifs, err := utils.DtoErrorSlice(notifs, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", serverrors.NewRepositoryError(err))
	}

	return domainNotifs, nil
}

func (er *EventsRepository) Get(ctx context.Context, id int) (domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	event, err := er.q.GetEvent(ctx, tx, int32(id))
	if err != nil {
		return domains.Event{}, fmt.Errorf("get event: %w", serverrors.NewRepositoryError(err))
	}

	return er.dto(event)
}

func (er *EventsRepository) GetLatest(ctx context.Context, taskdID int, taskType domains.TaskType) (domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	dbTaskType, err := er.repoTaskType(taskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("repo task type: %w", err)
	}

	event, err := er.q.GetLatestEvent(ctx, tx, goqueries.GetLatestEventParams{
		TaskID:   int32(taskdID),
		TaskType: dbTaskType,
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf("get latest event: %w", serverrors.NewRepositoryError(err))
	}

	return er.dto(event)
}

func (er *EventsRepository) Update(ctx context.Context, event domains.Event) error {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	_, err := er.q.UpdateEvent(ctx, tx, goqueries.UpdateEventParams{
		Text:          event.Text,
		NextSendTime:  pgxconv.Timestamptz(event.NextSendTime),
		FirstSendTime: pgxconv.Timestamptz(event.FirstSendTime),
		Done:          event.Done,
		ID:            int32(event.ID),
	})
	if err != nil {
		return fmt.Errorf("update event: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func (er *EventsRepository) Delete(ctx context.Context, id int) error {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	ns, err := er.q.DeleteEvent(ctx, tx, int32(id))
	if err != nil {
		return fmt.Errorf("delete event: %w", serverrors.NewRepositoryError(err))
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete event: %w", serverrors.NewNoDeletionsError("event"))
	}

	return nil
}

func (er *EventsRepository) ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	events, err := er.q.ListNotSendedEvents(ctx, tx, pgxconv.Timestamptz(till))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	domainEvents, err := utils.DtoErrorSlice(events, er.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return domainEvents, nil
}

func (er *EventsRepository) GetNearest(ctx context.Context) (domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	event, err := er.q.GetNearestEvent(ctx, tx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.Event{}, fmt.Errorf("get nearest event: %w", serverrors.NewNotFoundError(err, "event"))
		}

		return domains.Event{}, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return er.dto(event)
}
