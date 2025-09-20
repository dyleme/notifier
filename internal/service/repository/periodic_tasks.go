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

type PeriodicTaskRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewPeriodicTaskRepository(getter *txmanager.Getter) *PeriodicTaskRepository {
	return &PeriodicTaskRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

func (p *PeriodicTaskRepository) dtoWithTags(pt goqueries.PeriodicTask, tags []goqueries.Tag) (domain.PeriodicTask, error) {
	notifParams, err := parseNotificationParams(pt.NotificationParams)
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("parse notification params: %w", err)
	}
	task := domain.PeriodicTask{
		ID:                 int(pt.ID),
		Text:               pt.Text,
		Description:        pt.Description.String,
		UserID:             int(pt.UserID),
		Start:              pt.Start.Sub(time.Time{}),
		SmallestPeriod:     time.Duration(pt.SmallestPeriod) * time.Minute,
		BiggestPeriod:      time.Duration(pt.BiggestPeriod) * time.Minute,
		NotificationParams: notifParams,
		Tags:               slice.Dto(tags, dtoTag),
	}

	return task, nil
}

func (p *PeriodicTaskRepository) Add(ctx context.Context, task domain.PeriodicTask) (domain.PeriodicTask, error) {
	tx := p.getter.GetTx(ctx)
	pt, err := p.q.AddPeriodicTask(ctx, tx, goqueries.AddPeriodicTaskParams{
		UserID:             int64(task.UserID),
		Text:               task.Text,
		Start:              time.Time{}.Add(task.Start),
		SmallestPeriod:     int64(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int64(task.BiggestPeriod / time.Minute),
		Description:        sqlconv.NullableString(task.Description),
		NotificationParams: task.NotificationParams.JSON(),
	})
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("add periodic task: %w", err)
	}

	for _, t := range task.Tags {
		err = p.q.AddTagsToSmth(ctx, tx, goqueries.AddTagsToSmthParams{
			SmthID: pt.ID,
			TagID:  int64(t.ID),
			UserID: int64(t.UserID),
		})
	}
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("add tags to periodic task: %w", err)
	}

	pr, err := p.Get(ctx, int(pt.ID))
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("get periodic task: %w", err)
	}

	return pr, nil
}

func (p *PeriodicTaskRepository) Get(ctx context.Context, taskID int) (domain.PeriodicTask, error) {
	tx := p.getter.GetTx(ctx)
	task, err := p.q.GetPeriodicTask(ctx, tx, int64(taskID))
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("get periodic task: %w", err)
	}

	tags, err := p.q.ListTagsForSmth(ctx, tx, int64(taskID))
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("list tags for periodic task: %w", err)
	}

	return p.dtoWithTags(task, tags)
}

func (p *PeriodicTaskRepository) Update(ctx context.Context, task domain.PeriodicTask) error {
	op := "PeriodicTaskRepository.Update: %w"

	tx := p.getter.GetTx(ctx)
	pt, err := p.q.UpdatePeriodicTask(ctx, tx, goqueries.UpdatePeriodicTaskParams{
		ID:                 int64(task.ID),
		UserID:             int64(task.UserID),
		Start:              time.Time{}.Add(task.Start),
		Text:               task.Text,
		Description:        sqlconv.NullableString(task.Description),
		NotificationParams: task.NotificationParams.JSON(),
		SmallestPeriod:     int64(task.SmallestPeriod / time.Minute),
		BiggestPeriod:      int64(task.BiggestPeriod / time.Minute),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = syncTags(ctx, tx, p.q, task.UserID, int(pt.ID), task.Tags)
	if err != nil {
		return fmt.Errorf("sync tags: %w", err)
	}

	return nil
}

func (p *PeriodicTaskRepository) Delete(ctx context.Context, taskID int) error {
	op := "PeriodicTaskRepository.Delete: %w"

	tx := p.getter.GetTx(ctx)
	evs, err := p.q.DeletePeriodicTask(ctx, tx, int64(taskID))
	if err != nil {
		return fmt.Errorf(op, err)
	}
	if len(evs) == 0 {
		return fmt.Errorf(op, apperr.ErrNotFound)
	}

	err = p.q.DeleteAllTagsForSmth(ctx, tx, int64(taskID))
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}

func (p *PeriodicTaskRepository) List(ctx context.Context, userID int, params service.ListFilterParams) ([]domain.PeriodicTask, error) {
	tx := p.getter.GetTx(ctx)

	dbRows, err := p.q.ListPeriodicTasks(ctx, tx, goqueries.ListPeriodicTasksParams{
		UserID: int64(userID),
		Off:    int64(params.ListParams.Offset),
		Lim:    int64(params.ListParams.Limit),
		TagIds: slice.Dto(params.TagIDs, func(id int) int32 { return int32(id) }),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list periodic tasks: %w", err)
	}

	tasksIDs := slice.Dto(dbRows, func(t goqueries.ListPeriodicTasksRow) int64 { return t.PeriodicTask.ID })

	rows, err := listTagsForSmths(ctx, tx, p.q, tasksIDs)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("list tags for smths: %w", err)
		}
	}

	tasks := make([]domain.PeriodicTask, 0, len(dbRows))
	for _, pt := range dbRows {
		var tags []goqueries.Tag
		for _, row := range rows {
			if row.SmthID == pt.PeriodicTask.ID {
				tags = append(tags, row.Tag)
			}
		}
		task, err := p.dtoWithTags(pt.PeriodicTask, tags)
		if err != nil {
			return nil, fmt.Errorf("dto with tags: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
