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

func (er *EventsRepository) dto(dbEv goqueries.Event) (domains.Event, error) {
	taskType, err := er.domainTaskType(dbEv.TaskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("domain task type: %w", err)
	}

	event := domains.Event{
		ID:                 int(dbEv.ID),
		UserID:             int(dbEv.UserID),
		Text:               dbEv.Text,
		Description:        pgxconv.String(dbEv.Description),
		TaskType:           taskType,
		TaskID:             int(dbEv.TaskID),
		NotificationParams: dbEv.NotificationParams,
		NextSend:           pgxconv.TimeWithZone(dbEv.NextSend),
		FirstSend:          pgxconv.TimeWithZone(dbEv.FirstSend),
		Done:               dbEv.Done,
		Tags:               nil,
		Notify:             dbEv.Notify,
	}

	return event, nil
}

func (er *EventsRepository) dtoWithTags(dbEv goqueries.Event, dbTags []goqueries.Tag) (domains.Event, error) {
	taskType, err := er.domainTaskType(dbEv.TaskType)
	if err != nil {
		return domains.Event{}, fmt.Errorf("domain task type: %w", err)
	}

	event := domains.Event{
		ID:                 int(dbEv.ID),
		UserID:             int(dbEv.UserID),
		Text:               dbEv.Text,
		Description:        pgxconv.String(dbEv.Description),
		TaskType:           taskType,
		TaskID:             int(dbEv.TaskID),
		NotificationParams: dbEv.NotificationParams,
		NextSend:           pgxconv.TimeWithZone(dbEv.NextSend),
		FirstSend:          pgxconv.TimeWithZone(dbEv.FirstSend),
		Done:               dbEv.Done,
		Tags:               utils.DtoSlice(dbTags, dtoTag),
		Notify:             dbEv.Notify,
	}

	return event, nil
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
		NextSend:               pgxconv.Timestamptz(event.NextSend),
		NotificationParams: event.NotificationParams,
	})
	if err != nil {
		return domains.Event{}, fmt.Errorf("add event: %w", serverrors.NewRepositoryError(err))
	}

	_, err = er.q.AddTagsToSmth(ctx, tx, utils.DtoSlice(event.Tags, func(t domains.Tag) goqueries.AddTagsToSmthParams {
		return goqueries.AddTagsToSmthParams{
			SmthID: ev.ID,
			TagID:  int32(t.ID),
			UserID: int32(event.UserID),
		}
	}))
	if err != nil {
		return domains.Event{}, fmt.Errorf("add tags to smth: %w", serverrors.NewRepositoryError(err))
	}

	return er.Get(ctx, int(ev.ID))
}

func (er *EventsRepository) List(ctx context.Context, userID int, params service.ListEventsFilterParams) ([]domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	rowsEvents, err := er.q.ListUserEvents(ctx, tx, goqueries.ListUserEventsParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamptz(params.TimeBorders.From),
		ToTime:   pgxconv.Timestamptz(params.TimeBorders.To),
		TagIds:   utils.DtoSlice(params.Tags, func(tagID int) int32 { return int32(tagID) }),
		Off:      int32(params.ListParams.Offset),
		Lim:      int32(params.ListParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list user events: %w", serverrors.NewRepositoryError(err))
	}

	tasksIDs := utils.DtoSlice(rowsEvents, func(t goqueries.ListUserEventsRow) int32 { return t.Event.ID })

	rows, err := er.q.ListTagsForSmths(ctx, tx, tasksIDs)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("list tags for smths: %w", serverrors.NewRepositoryError(err))
		}
	}

	events := make([]domains.Event, 0, len(rowsEvents))
	for _, ev := range rowsEvents {
		var tags []goqueries.Tag
		for _, row := range rows {
			if row.SmthID == ev.Event.ID {
				tags = append(tags, row.Tag)
			}
		}
		event, err := er.dtoWithTags(ev.Event, tags)
		if err != nil {
			return nil, fmt.Errorf("list user events: %w", serverrors.NewServiceError(err))
		}
		events = append(events, event)
	}

	return events, nil
}

func (er *EventsRepository) Get(ctx context.Context, id int) (domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	event, err := er.q.GetEvent(ctx, tx, int32(id))
	if err != nil {
		return domains.Event{}, fmt.Errorf("get event: %w", serverrors.NewRepositoryError(err))
	}

	tags, err := er.q.ListTagsForSmth(ctx, tx, event.ID)
	if err != nil {
		return domains.Event{}, fmt.Errorf("list tags for smth: %w", serverrors.NewRepositoryError(err))
	}

	return er.dtoWithTags(event, tags)
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

	tags, err := er.q.ListTagsForSmth(ctx, tx, event.ID)
	if err != nil {
		return domains.Event{}, fmt.Errorf("list tags for smth: %w", serverrors.NewRepositoryError(err))
	}

	return er.dtoWithTags(event, tags)
}

func (er *EventsRepository) Update(ctx context.Context, event domains.Event) error {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	_, err := er.q.UpdateEvent(ctx, tx, goqueries.UpdateEventParams{
		Text:      event.Text,
		NextSend:      pgxconv.Timestamptz(event.NextSend),
		FirstSend: pgxconv.Timestamptz(event.FirstSend),
		Done:      event.Done,
		ID:        int32(event.ID),
	})
	if err != nil {
		return fmt.Errorf("update event: %w", serverrors.NewRepositoryError(err))
	}

	err = syncTags(ctx, tx, er.q, event.UserID, event.ID, event.Tags)
	if err != nil {
		return fmt.Errorf("sync tags: %w", err)
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

func (er *EventsRepository) GetNearest(ctx context.Context) (time.Time, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)

	t, err := er.q.GetNearestEventTime(ctx, tx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf("get nearest event: %w", serverrors.NewNotFoundError(err, "event"))
		}

		return time.Time{}, fmt.Errorf("list not sended notifiations: %w", serverrors.NewRepositoryError(err))
	}

	return pgxconv.TimeWithZone(t), nil
}

func (er *EventsRepository) ListDayEvents(ctx context.Context, userID, timeZoneOffset int) ([]domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	events, err := er.q.ListUserDailyEvents(ctx, tx, goqueries.ListUserDailyEventsParams{
		UserID:     int32(userID),
		TimeOffset: pgxconv.Timestamptz(time.Time{}.Add(time.Duration(timeZoneOffset) * time.Hour)),
	})
	if err != nil {
		return nil, fmt.Errorf("list user daily events: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoErrorSlice(events, er.dto)
}

func (er *EventsRepository) ListNotDoneEvents(ctx context.Context, userID int) ([]domains.Event, error) {
	tx := er.getter.DefaultTrOrDB(ctx, er.db)
	events, err := er.q.ListNotDoneEvents(ctx, tx, int32(userID))
	if err != nil {
		return nil, fmt.Errorf("list not done events: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoErrorSlice(events, er.dto)
}
