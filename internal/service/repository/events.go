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

type EventsRepository struct {
	q      *queries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func (r *Repository) Events() service.EventsRepository {
	return r.eventsRepository
}

const UnkownTaskType = queries.TaskType("unknown_type")

func (*EventsRepository) repoTaskType(taskType domains.TaskType) (queries.TaskType, error) {
	switch taskType {
	case domains.PeriodicTaskType:
		return queries.TaskTypePeriodicTask, nil
	case domains.BasicTaskType:
		return queries.TaskTypeBasicTask, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (*EventsRepository) domainTaskType(taskType queries.TaskType) (domains.TaskType, error) {
	switch taskType {
	case queries.TaskTypePeriodicTask:
		return domains.PeriodicTaskType, nil
	case queries.TaskTypeBasicTask:
		return domains.BasicTaskType, nil
	default:
		return "", serverrors.NewBusinessLogicError("unknown task type")
	}
}

func (n *EventsRepository) dto(event queries.Event) (domains.Event, error) {
	taskType, err := n.domainTaskType(event.TaskType)
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
		SendTime:           pgxconv.TimeWithZone(event.SendTime),
		Sended:             event.Sended,
		Done:               event.Done,
	}, nil
}

func (n *EventsRepository) Add(ctx context.Context, event domains.Event) (domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	taskType, err := n.repoTaskType(event.TaskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("repo task type: %w", err)
	}
	ev, err := n.q.AddEvent(ctx, tx, queries.AddEventParams{
		UserID:   int32(event.UserID),
		Text:     event.Text,
		TaskID:   int32(event.TaskID),
		TaskType: taskType,
		SendTime: pgxconv.Timestamptz(event.SendTime),
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf("add event: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(ev)
}

func (n *EventsRepository) List(ctx context.Context, userID int, timeBorderes timeborders.TimeBorders, listParams service.ListParams) ([]domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	notifs, err := n.q.ListUserEvents(ctx, tx, queries.ListUserEventsParams{
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

	domainNotifs, err := utils.DtoErrorSlice(notifs, n.dto)
	if err != nil {
		return nil, fmt.Errorf("list user events: %w", serverrors.NewRepositoryError(err))
	}

	return domainNotifs, nil
}

func (n *EventsRepository) Get(ctx context.Context, id int) (domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	event, err := n.q.GetEvent(ctx, tx, int32(id))
	if err != nil {
		return domains.Event{}, fmt.Errorf("get event: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(event)
}

func (n *EventsRepository) GetLatest(ctx context.Context, taskdID int, taskType domains.TaskType) (domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)
	dbTaskType, err := n.repoTaskType(taskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("repo task type: %w", err)
	}

	event, err := n.q.GetLatestEvent(ctx, tx, queries.GetLatestEventParams{
		TaskID:   int32(taskdID),
		TaskType: dbTaskType,
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf("get latest event: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(event)
}

func (n *EventsRepository) Update(ctx context.Context, event domains.Event) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	_, err := n.q.UpdateEvent(ctx, tx, queries.UpdateEventParams{
		ID:       int32(event.ID),
		Text:     event.Text,
		SendTime: pgxconv.Timestamptz(event.SendTime),
		Sended:   event.Sended,
		Done:     event.Done,
	})
	if err != nil {
		return fmt.Errorf("update event: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func (n *EventsRepository) Delete(ctx context.Context, id int) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	ns, err := n.q.DeleteEvent(ctx, tx, int32(id))
	if err != nil {
		return fmt.Errorf("delete event: %w", serverrors.NewRepositoryError(err))
	}

	if len(ns) == 0 {
		return fmt.Errorf("delete event: %w", serverrors.NewNoDeletionsError("event"))
	}

	return nil
}

func (n *EventsRepository) ListNotSended(ctx context.Context, till time.Time) ([]domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	events, err := n.q.ListNotSendedEvents(ctx, tx, pgxconv.Timestamptz(till))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	domainEvents, err := utils.DtoErrorSlice(events, n.dto)
	if err != nil {
		return nil, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return domainEvents, nil
}

func (n *EventsRepository) GetNearest(ctx context.Context, till time.Time) (domains.Event, error) {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	event, err := n.q.GetNearestEvent(ctx, tx, pgxconv.Timestamptz(till))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.Event{}, fmt.Errorf("get nearest event: %w", serverrors.NewNotFoundError(err, "event"))
		}

		return domains.Event{}, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return n.dto(event)
}

func (n *EventsRepository) MarkSended(ctx context.Context, ids []int) error {
	tx := n.getter.DefaultTrOrDB(ctx, n.db)

	int32IDs := utils.DtoSlice(ids, func(i int) int32 { return int32(i) })
	err := n.q.MarkSendedNotifiatoins(ctx, tx, int32IDs)
	if err != nil {
		return fmt.Errorf("mark sended events: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}
