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

type PeriodicTaskRepository struct {
	q      *goqueries.Queries
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPeriodicTaskRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *PeriodicTaskRepository {
	return &PeriodicTaskRepository{
		q:      goqueries.New(),
		pool:   db,
		getter: getter,
	}
}

func (p *PeriodicTaskRepository) dtoWithTags(pt goqueries.PeriodicTask, tags []goqueries.Tag) domains.PeriodicTask {
	task := domains.PeriodicTask{
		ID:                 int(pt.ID),
		Text:               pt.Text,
		Description:        pgxconv.String(pt.Description),
		UserID:             int(pt.UserID),
		Start:              pgxconv.TimeWithZone(pt.Start).Sub(time.Time{}),
		SmallestPeriod:     time.Duration(pt.SmallestPeriod) * time.Minute,
		BiggestPeriod:      time.Duration(pt.BiggestPeriod) * time.Minute,
		NotificationParams: pt.NotificationParams,
		Tags:               utils.DtoSlice(tags, dtoTag),
		Notify:             pt.Notify,
	}

	return task
}

func (p *PeriodicTaskRepository) Add(ctx context.Context, task domains.PeriodicTask) (domains.PeriodicTask, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.pool)
	pt, err := p.q.AddPeriodicTask(ctx, tx, goqueries.AddPeriodicTaskParams{
		UserID:             int32(task.UserID),
		Text:               task.Text,
		Start:              pgxconv.Timestamptz(time.Time{}.Add(task.Start)),
		SmallestPeriod:     int32(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(task.BiggestPeriod / time.Minute),
		Description:        pgxconv.Text(task.Description),
		NotificationParams: task.NotificationParams,
	})
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("add periodic task: %w", serverrors.NewRepositoryError(err))
	}

	_, err = p.q.AddTagsToSmth(ctx, tx, utils.DtoSlice(task.Tags, func(tag domains.Tag) goqueries.AddTagsToSmthParams {
		return goqueries.AddTagsToSmthParams{
			SmthID: pt.ID,
			TagID:  int32(tag.ID),
			UserID: int32(task.UserID),
		}
	}))
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("add tags to periodic task: %w", serverrors.NewRepositoryError(err))
	}

	pr, err := p.Get(ctx, int(pt.ID))
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("get periodic task: %w", serverrors.NewRepositoryError(err))
	}

	return pr, nil
}

func (p *PeriodicTaskRepository) Get(ctx context.Context, taskID int) (domains.PeriodicTask, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.pool)
	task, err := p.q.GetPeriodicTask(ctx, tx, int32(taskID))
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("get periodic task: %w", serverrors.NewRepositoryError(err))
	}

	tags, err := p.q.ListTagsForSmth(ctx, tx, int32(taskID))
	if err != nil {
		return domains.PeriodicTask{}, fmt.Errorf("list tags for periodic task: %w", serverrors.NewRepositoryError(err))
	}

	return p.dtoWithTags(task, tags), nil
}

func (p *PeriodicTaskRepository) Update(ctx context.Context, task domains.PeriodicTask) error {
	op := "PeriodicTaskRepository.Update: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.pool)
	pt, err := p.q.UpdatePeriodicTask(ctx, tx, goqueries.UpdatePeriodicTaskParams{
		ID:                 int32(task.ID),
		UserID:             int32(task.UserID),
		Start:              pgxconv.Timestamptz(time.Time{}.Add(task.Start)),
		Text:               task.Text,
		Description:        pgxconv.Text(task.Description),
		NotificationParams: task.NotificationParams,
		SmallestPeriod:     int32(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int32(task.BiggestPeriod / time.Minute),
	})
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	err = syncTags(ctx, tx, p.q, task.UserID, int(pt.ID), task.Tags)
	if err != nil {
		return fmt.Errorf("sync tags: %w", err)
	}

	return nil
}

func (p *PeriodicTaskRepository) Delete(ctx context.Context, taskID int) error {
	op := "PeriodicTaskRepository.Delete: %w"

	tx := p.getter.DefaultTrOrDB(ctx, p.pool)
	evs, err := p.q.DeletePeriodicTask(ctx, tx, int32(taskID))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("periodic task"))
	}

	err = p.q.DeleteAllTagsForSmth(ctx, tx, int32(taskID))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (p *PeriodicTaskRepository) List(ctx context.Context, userID int, params service.ListFilterParams) ([]domains.PeriodicTask, error) {
	tx := p.getter.DefaultTrOrDB(ctx, p.pool)

	dbRows, err := p.q.ListPeriodicTasks(ctx, tx, goqueries.ListPeriodicTasksParams{
		UserID: int32(userID),
		Off:    int32(params.ListParams.Offset),
		Lim:    int32(params.ListParams.Limit),
		TagIds: utils.DtoSlice(params.TagIDs, func(id int) int32 { return int32(id) }),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list periodic tasks: %w", serverrors.NewRepositoryError(err))
	}

	tasksIDs := utils.DtoSlice(dbRows, func(t goqueries.ListPeriodicTasksRow) int32 { return t.PeriodicTask.ID })

	rows, err := p.q.ListTagsForSmths(ctx, tx, tasksIDs)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("list tags for smths: %w", serverrors.NewRepositoryError(err))
		}
	}

	tasks := make([]domains.PeriodicTask, 0, len(dbRows))
	for _, pt := range dbRows {
		var tags []goqueries.Tag
		for _, row := range rows {
			if row.SmthID == pt.PeriodicTask.ID {
				tags = append(tags, row.Tag)
			}
		}
		tasks = append(tasks, p.dtoWithTags(pt.PeriodicTask, tags))
	}

	return tasks, nil
}
